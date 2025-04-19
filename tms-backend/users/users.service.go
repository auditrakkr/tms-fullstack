package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	auth_dto "github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/auditrakkr/tms-fullstack/tms-backend/search"
	"github.com/gin-gonic/gin"

	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo                 repositories.Repository[models.User]
	roleRepo                 repositories.Repository[models.Role]
	tenantRepo               repositories.Repository[models.Tenant]
	tenantTeamRepo           repositories.Repository[models.TenantTeam]
	tenantAccountOfficerRepo repositories.Repository[models.TenantAccountOfficer]
	usersSearchService       *search.UsersSearchService
	db                       *gorm.DB
}

func NewUserService() *UserService {
	// Initialize the user search service
	searchService := search.NewUsersSearchService()
	ctx := context.Background()
	err := searchService.Initialize(ctx)
	if err != nil {
		fmt.Printf("Error initializing search service: %v\n", err)
	}
	return &UserService{
		userRepo:                 repositories.Repository[models.User]{DB: database.DB},
		roleRepo:                 repositories.Repository[models.Role]{DB: database.DB},
		tenantRepo:               repositories.Repository[models.Tenant]{DB: database.DB},
		tenantTeamRepo:           repositories.Repository[models.TenantTeam]{DB: database.DB},
		tenantAccountOfficerRepo: repositories.Repository[models.TenantAccountOfficer]{DB: database.DB},
		usersSearchService:       searchService,
		db:                       database.DB,
	}
}

/* Create  */

func (s *UserService) CreateUser(c *gin.Context, createUserDto *dto.CreateUserDto) (*models.User, error) {
	newUser := &models.User{}
	if err := copier.Copy(newUser, createUserDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	newUser, err := s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	newUser.PasswordHash = string(hashedPassword)

	user, err := s.userRepo.Save(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	//add to elastic search
	if s.usersSearchService != nil {
		ctx := context.Background()
		err = s.usersSearchService.IndexUser(ctx, *user)
		if err != nil {
			return nil, fmt.Errorf("failed to add user to search index: %v", err)
		}
	}

	// Send confirmation email if enabled
	if global.AUTO_SEND_CONFIRM_EMAIL {
		s.ConfirmEmailRequest(&createUserDto.PrimaryEmailAddress, 0, true, c)
	}

	return user, nil
}

/*UPDATE section  */
func (s *UserService) Update(userId uint, updateUserDto *dto.UpdateUserDto) (*models.User, error) {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	if err := copier.Copy(user, updateUserDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)

	}

	// Update the user in the search index
	if s.usersSearchService != nil {
		ctx := context.Background()
		err = s.usersSearchService.Update(ctx, *user)
		if err != nil {
			return nil, fmt.Errorf("failed to update user in search index: %v", err)
		}
	}

	return user, nil
}

/* Delete */

func (s *UserService) DeleteUser(userId uint) error {

	// Remove from search index first
	if s.usersSearchService != nil {
		ctx := context.Background()
		err := s.usersSearchService.Remove(ctx, int(userId))
		if err != nil {
			fmt.Printf("Warning: failed to remove user from search index: %v\n", err)
		}
	}

	// Delete the user from the database
	err := s.userRepo.Delete(userId)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

func (s *UserService) RemoveUser(user *models.User) (*models.User, error) {
	removedUser, err := s.userRepo.Remove(user)
	if err != nil {
		return nil, fmt.Errorf("failed to remove user: %v", err)
	}
	return &removedUser, nil
}

/* Read Section */

func (s *UserService) FindAllWithOptions(findOptions map[string]any) ([]models.User, int64, error) {
	users, totalCount, err := s.userRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find all users with options: %v", err)
	}
	return users, totalCount, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to find all users: %v", err)
	}
	return users, nil
}

/* func (s *UserService) FindAll(findOptions map[string]any) ([]models.User, int64, error) {
	users, totalCount, err := s.userRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find all users with options: %v", err)
	}
	return users, totalCount, nil
} */

func (s *UserService) FindOne(userId uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) FindByPrimaryEmailAddress(primaryEmailAddress string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("primary_email_address = ?", primaryEmailAddress).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by primary email address: %v", err)
	}
	if user.PrimaryEmailAddress == "" { // Assuming PrimaryEmailAddress is a required field
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *UserService) FindById(id uint) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		// Use Select if you need to explicitly select the RefreshTokenHash
		Select("*"). // This selects all fields including RefreshTokenHash
		Preload("Roles").
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("there was a problem getting user: %v", err)
	}

	if user.ID == 0 { // Check if user was found
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

func (s *UserService) FindByGoogleId(googleId string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Select("*"). // Select all fields including refresh_token_hash
		Preload("Roles").
		Preload("GoogleProfile").
		Joins("JOIN google_profiles ON google_profiles.user_id = users.id").
		Where("google_profiles.google_id = ?", googleId).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("there was a problem getting user: %v", err)
	}

	if user.ID == 0 { // Check if user was found
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

func (s *UserService) FindByFacebookId(facebookId string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Select("*"). // Select all fields including refresh_token_hash
		Preload("Roles").
		Preload("FacebookProfile").
		Joins("JOIN facebook_profiles ON facebook_profiles.user_id = users.id").
		Where("facebook_profiles.facebook_id = ?", facebookId).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("there was a problem getting user: %v", err)
	}

	if user.ID == 0 { // Check if user was found
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

/* Invoked to setRefreshTokenHash after successful login. */
func (s *UserService) SetRefreshTokenHash(userId uint, refreshTokenHash string) error {
	// Hash the refresh token
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash refresh token: %v", err)
	}

	// Update only the RefreshTokenHash field
	err = s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Update("refresh_token_hash", string(hashedRefreshToken)).Error
	if err != nil {
		return fmt.Errorf("failed to update refresh token hash: %v", err)
	}

	return nil
}

func (s *UserService) SetGoogleProfile(userId uint, googleProfile *auth_dto.GoogleProfileDto) (*models.User, error) {
	newGoogleProfile := &models.GoogleProfile{}
	if err := copier.Copy(newGoogleProfile, googleProfile); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	newGoogleProfile.UserID = userId
	err := s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("GoogleProfile").Replace(newGoogleProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to set google profile: %v", err)
	}
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) SetFacebookProfile(userId uint, facebookProfile *auth_dto.FacebookProfileDto) (*models.User, error) {
	newFacebookProfile := &models.FacebookProfile{}
	if err := copier.Copy(newFacebookProfile, facebookProfile); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	newFacebookProfile.UserID = userId
	err := s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("FacebookProfile").Replace(newFacebookProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to set facebook profile: %v", err)
	}
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// SetUserPassword sets a user's password
func (s *UserService) SetUserPassword(userId uint, password string) (bool, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %v", err)
	}

	// Update the user's password
	err = s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Update("password_hash", string(hashedPassword)).Error
	if err != nil {
		return false, fmt.Errorf("failed to update user password: %v", err)
	}

	return true, nil
}

// SetUserPhoto handles uploading user's profile photo
func (s *UserService) SetUserPhoto(userId uint, file *multipart.FileHeader, c *gin.Context) error {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	// check file size
	if file.Size > global.PHOTO_FILE_SIZE_LIMIT {
		return fmt.Errorf("file too large (max %d bytes)", global.PHOTO_FILE_SIZE_LIMIT)
	}

	// generate unique filename
	var fileName string
	if user.Photo != "" {
		fileName = user.Photo
	} else {
		fileName = fmt.Sprintf("%d_%s", user.ID, file.Filename)
	}

	//Ensure the upload directory exists
	uploadDir := filepath.Join(global.UPLOAD_DIRECTORY, "photos")
	if err :=  os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Save the file
	dest := filepath.Join(uploadDir, fileName)
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Update user record
	user.Photo = fileName
	user.PhotoMimeType = file.Header.Get("Content-Type")
	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to update user photo info: %v", err)
	}

	return nil

}

// GetUserPhoto serves a user's photo
func (s *UserService) GetUserPhoto(userId uint, c *gin.Context) error {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	// If no photo, use default
	if user.Photo == "" {
		c.File(filepath.Join(global.UPLOAD_DIRECTORY, "photos", "blankPhotoAvatar.png"))
		return nil
	}

	// Serve photo file
	photoPath := filepath.Join(global.UPLOAD_DIRECTORY, "photos", user.Photo)
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		// File doesn't exist, use default
		c.File(filepath.Join(global.UPLOAD_DIRECTORY, "photos", "blankPhotoAvatar.png"))
		return nil
	}

	c.Header("Content-Type", user.PhotoMimeType)
	c.File(photoPath)
	return nil
}

// generateRandomToken generates a random token for password reset or email verification
func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}


// ResetPasswordRequest sends a password reset link to the user's email
func (s *UserService) ResetPasswordRequest(email string, c *gin.Context) (*global.GenericNotificationResponse, error) {
	user, err := s.FindByPrimaryEmailAddress(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	if user == nil {
		return &global.GenericNotificationResponse{
			NotificationClass:   "is-success",
			NotificationMessage: fmt.Sprintf("If your email %s is found, you will receive email shortly for password reset", email),
		}, nil
	}

	// Generate reset token
	token, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Set token and expiration
	user.ResetPasswordToken = token
	user.ResetPasswordExpiration = time.Now().Add(global.PASSWORD_RESET_EXPIRATION)

	// Save user with new token
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to save reset token: %v", err)
	}

	// Build reset URL with proper protocol and URL structure
    var globalPrefixUrl string
    if global.USE_API_VERSION_IN_URL {
        globalPrefixUrl = fmt.Sprintf("/%s", global.API_VERSION)
    } else {
        globalPrefixUrl = ""
    }

    // Use protocol from request or fallback to configured protocol
    protocol := c.Request.URL.Scheme
    if protocol == "" {
        protocol = global.PROTOCOL // Fallback to default protocol
    }

    resetURL := fmt.Sprintf("%s://%s%s/users/reset-password/%s",
        protocol,
        c.Request.Host,
        globalPrefixUrl,
        token)

	// Prepare email content
	mailText := strings.Replace(global.ResetPasswordMailOptionSettings.TextTemplate, "{url}", resetURL, -1)

	// Configure mail options
    mailOptions := global.MailOptions{
        To:      user.PrimaryEmailAddress,
        From:    global.ResetPasswordMailOptionSettings.From,
        Subject: global.ResetPasswordMailOptionSettings.Subject,
        Text:    mailText,
    }

	// Send email asynchronously
    global.SendMailAsync(mailOptions)

    return &global.GenericNotificationResponse{
        NotificationClass:   "is-success",
        NotificationMessage: fmt.Sprintf("If your email %s is found, you will receive email shortly for password reset", email),
    }, nil
}

/* // ResetPassword handles the password reset process
func (s *UserService) ResetPassword(token string, newPassword *string,  c *gin.Context)  error {
	// Find user by reset token
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("reset_password_token = ?", token).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid token")
		}
		return fmt.Errorf("error finding user: %v", err)
	}

	// Check token expiration
	if user.ResetPasswordExpiration.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	if newPassword != nil {
		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}

		// Update user
		user.PasswordHash = string(hashedPassword)
		user.ResetPasswordToken = ""
		user.ResetPasswordExpiration = time.Time{} // Clear expiration

		err = s.userRepo.Update(&user)
		if err != nil {
			return fmt.Errorf("failed to update user password: %v", err)
		}

		return nil
	}
} */


// ConfirmEmailRequest sends an email verification link
func (s *UserService) ConfirmEmailRequest(email *string, userId uint, isPrimary bool, c *gin.Context) (*global.GenericNotificationResponse, error) {
	var user *models.User
	var err error

	// Find user by ID or email
    if userId > 0 {
        user, err = s.userRepo.FindByID(userId)
    } else if email != nil {
        if isPrimary {
            user, err = s.FindByPrimaryEmailAddress(*email)
        } else {
            // Find by backup email address
            var foundUser models.User
            err = s.userRepo.CreateQueryBuilder().
                Where("backup_email_address = ?", *email).
                First(&foundUser).Error
            if err == nil {
                user = &foundUser
            }
        }
    } else {
        return nil, fmt.Errorf("either email or userId must be provided")
    }
	if err != nil {
        return nil, fmt.Errorf("error finding user: %v", err)
    }

	if user == nil {
        // Don't reveal if user exists
        return &global.GenericNotificationResponse{
            NotificationClass:   "is-info",
            NotificationMessage: "If valid user, you will receive email shortly for verification",
        }, nil
    }

	// Generate verification token
    token, err := generateRandomToken(32)
    if err != nil {
        return nil, fmt.Errorf("failed to generate token: %v", err)
    }

	if isPrimary {
        user.PrimaryEmailVerificationToken = token
    } else {
        user.BackupEmailVerificationToken = token
    }

	// Set expiration time for the token
	user.EmailVerificationTokenExpiration = time.Now().Add(global.EMAIL_VERIFICATION_EXPIRATION)

	// Save the user with the new token
	_, err = s.userRepo.Save(user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	// Build verification URL
	var endpoint string
	if isPrimary {
		endpoint = "confirm-primary-email"
	} else {
		endpoint = "confirm-primary-email"
	}


	// Get API version prefix if needed
    var globalPrefixUrl string
    if global.USE_API_VERSION_IN_URL {
        globalPrefixUrl = fmt.Sprintf("/%s", global.API_VERSION)
    } else {
        globalPrefixUrl = ""
    }

	// Determine protocol
    protocol := c.Request.URL.Scheme
    if protocol == "" {
        protocol = global.PROTOCOL // Fallback to default protocol
    }

	verificationURL := fmt.Sprintf("%s://%s%s/users/%s/%s",
        protocol,
        c.Request.Host,
        globalPrefixUrl,
        endpoint,
        token)

	// Prepare email content
	mailText := strings.Replace(global.ConfirmEmailMailOptionSettings.TextTemplate, "{url}", verificationURL, -1)
	//mailHTML := strings.Replace(global.ConfirmEmailMailOptionSettings.HtmlTemplate, "{url}", verificationURL, -1)



	// Determine recipient email address
	var recipientEmail string
	if isPrimary {
		recipientEmail = user.PrimaryEmailAddress
	} else {
		recipientEmail = user.BackupEmailAddress
	}

	// Configure mail options
    mailOptions := global.MailOptions{
        To:      recipientEmail,
    From:    global.ConfirmEmailMailOptionSettings.From,
    Subject: global.ConfirmEmailMailOptionSettings.Subject,
    Text:    mailText,
    //Html:    mailHTML,
    }

	// Send email asynchronously
    global.SendMailAsync(mailOptions)

    return &global.GenericNotificationResponse{
        NotificationClass:   "is-info",
        NotificationMessage: "If valid user, you will receive email shortly for verification",
    }, nil

}