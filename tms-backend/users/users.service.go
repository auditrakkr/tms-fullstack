package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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


/*Let's work on functions to set/add and unset/remove relations. set/unset applies to x-to-one and add/remove applies to x-to-many */
//1. Roles
// CreateAndAddRole creates a new role and assigns it to a user in a transaction
func (s *UserService) CreateAndAddRole(userId uint, createRoleDto *dto.CreateRoleDto) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %v", tx.Error)
	}

	// Defer either commit or rollback based on success/failure
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the new role
	newRole := &models.Role{}
	if err := copier.Copy(newRole, createRoleDto); err != nil {}

	// Save the new role in the transaction
    if err := tx.Create(newRole).Error; err != nil {
        tx.Rollback()
        // Check for unique constraint violation
        if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
            return fmt.Errorf("role with this name already exists: %v", err)
        }
        return fmt.Errorf("failed to create role: %v", err)
    }

	return nil

}

func (s *UserService) AddRoleById(userId uint, roleId uint) error {
	role := &models.Role{}
	if err := s.db.First(role, roleId).Error; err != nil {
		return fmt.Errorf("failed to find role: %v", err)
	}

	err := s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("Roles").
		Append(role)

	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role already assigned to user: %v", err)
		}
		return fmt.Errorf("failed to add role to user: %v", err)
	}
	return nil
}

func (s *UserService) AddRolesById(userId uint, roleIds []uint) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Where("id IN ?", roleIds).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %v", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found")
	}

	// Add associations
	err = s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("Roles").
		Append(roles)

	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("one or more roles already assigned to user: %v", err)
		}
		return nil, fmt.Errorf("failed to add roles to user: %v", err)
	}

	// Return the updated roles list
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return user.Roles, nil
}

func (s *UserService) RemoveRoleById(userId uint, roleId uint) ([]models.Role, error) {
	// Get the role to remove
	role := &models.Role{}
	if err := s.db.First(role, roleId).Error; err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}

	// Remove the association
	err := s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("Roles").
		Delete(role)

	if err != nil {
		return nil, fmt.Errorf("failed to remove role from user: %v", err)
	}

	// Remove the updated roles list
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return user.Roles, nil
}

func (s *UserService) RemoveRolesById(userId uint, roleIds []uint) ([]models.Role, error) {
	// Get the roles to remove
	var roles []models.Role
	err := s.db.Where("id IN ?", roleIds).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %v", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found")
	}

	// Remove the associations
	err = s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Association("Roles").
		Delete(roles)

	if err != nil {
		return nil, fmt.Errorf("failed to remove roles from user: %v", err)
	}

	// Return the updated roles list
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return user.Roles, nil

}

/*Some user perculiarities*/

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

func (s *UserService) FindByConfirmedPrimaryEmailAddress(primaryEmailAddress string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("primary_email_address = ? AND is_primary_email_verified = true", primaryEmailAddress).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by confirmed primary email address: %v", err)
	}
	if user.PrimaryEmailAddress == "" { // Assuming PrimaryEmailAddress is a required field
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *UserService) FindByResetPassToken(resetPasswordToken string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("reset_password_token = ?", resetPasswordToken).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by reset password token: %v", err)
	}

	return &user, nil
}

func (s *UserService) FindByPrimaryEmailVerificationToken(primaryEmailVerificationToken string) (*models.User, error) {
    var user models.User
    err := s.userRepo.CreateQueryBuilder().
        Where("primary_email_verification_token = ?", primaryEmailVerificationToken).
        First(&user).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // No user found but not an error
        }
        return nil, fmt.Errorf("there was a problem getting user: %v", err)
    }

    return &user, nil
}

func (s *UserService) FindByBackupEmailVerificationToken(backupEmailVerificationToken string) (*models.User, error) {
    var user models.User
    err := s.userRepo.CreateQueryBuilder().
        Where("backup_email_verification_token = ?", backupEmailVerificationToken).
        First(&user).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // No user found but not an error
        }
        return nil, fmt.Errorf("there was a problem getting user: %v", err)
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

// ResetPassword handles the password reset process
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
        // Token expired - render view with error message
        c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
            "title":                fmt.Sprintf("%s - Reset Password", global.APP_NAME),
            "sendForm":             false,
            "notificationVisibility": "",
            "notificationClass":    "is-danger",
            "notificationMessage":  "Invalid token: expired",
        })
        return nil
    }

	// If newPassword is provided, update user's password
    if newPassword != nil {
        // Hash the new password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
        if err != nil {
            return fmt.Errorf("failed to hash password: %v", err)
        }

        // Update user with new password and clear reset token fields
        user.PasswordHash = string(hashedPassword)
        user.ResetPasswordToken = ""
        user.ResetPasswordExpiration = time.Time{} // Clear expiration

        err = s.userRepo.Update(&user)
        if err != nil {
            return fmt.Errorf("failed to update user password: %v", err)
        }

        // Password successfully changed - render success view
        c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
            "title":                fmt.Sprintf("%s - Reset Password", global.APP_NAME),
            "sendForm":             false,
            "notificationVisibility": "",
            "notificationClass":    "is-success",
            "notificationMessage":  "New password successfully saved",
        })

        // Consider sending confirmation email here (optional)
        // s.sendPasswordChangedConfirmation(user.PrimaryEmailAddress, c)

        return nil
    }

	// No password provided yet, show the password reset form
    var globalPrefixUrl string
    if global.USE_API_VERSION_IN_URL {
        globalPrefixUrl = fmt.Sprintf("/%s", global.API_VERSION)
    } else {
        globalPrefixUrl = ""
    }

    returnUrl := fmt.Sprintf("%s/users/reset-password/%s", globalPrefixUrl, token)

    c.HTML(http.StatusOK, "users/reset-password.html", gin.H{
        "title":                fmt.Sprintf("%s - Reset Password", global.APP_NAME),
        "sendForm":             true,
        "returnUrl":            returnUrl,
        "notificationVisibility": "is-hidden",
    })

    return nil
}


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


// ConfirmEmail confirms an email address with a verification token
func (s *UserService) ConfirmEmail(token string, isPrimary bool, c *gin.Context) error {
	// Find user by verification token
	var user models.User
	var err error

	if isPrimary {
		err = s.userRepo.CreateQueryBuilder().
			Where("primary_email_verification_token = ?", token).
			First(&user).Error
	} else {
		err = s.userRepo.CreateQueryBuilder().
			Where("backup_email_verification_token = ?", token).
			First(&user).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid token")
		}
		return fmt.Errorf("error finding user: %v", err)
	}

	// Check token expiration
	if user.EmailVerificationTokenExpiration.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	if isPrimary {
		user.IsPrimaryEmailVerified = true
		user.PrimaryEmailVerificationToken = ""
	} else {
		user.IsBackupEmailVerified = true
		user.BackupEmailVerificationToken = ""
	}

	user.EmailVerificationTokenExpiration = time.Time{} // Clear expiration

	// Save user with updated verification status
	err = s.userRepo.Update(&user)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}


// SearchForUsers searches for users based on a query
func (s *UserService) SearchForUsers(text string, returnElasticSearchHitsDirectly bool) (interface{}, error) {
	ctx := context.Background()

	// Search for users matching the text
    results, err := s.usersSearchService.Search(ctx, text)
    if err != nil {
        return nil, fmt.Errorf("failed to search users: %v", err)
    }

	if returnElasticSearchHitsDirectly {
		return results, nil
	} else {
		// Extract IDs from search results
        var ids []uint
        for _, result := range results {
            // UserSearchBody has an ID field of type int, convert to uint
            ids = append(ids, uint(result.ID))
        }

		// If no results found, return empty array
        if len(ids) == 0 {
            return []models.User{}, nil
        }

		// Fetch actual users from the database using the IDs
        var users []models.User
        err = s.userRepo.CreateQueryBuilder().
            Where("id IN ?", ids).
            Find(&users).Error

        if err != nil {
            return nil, fmt.Errorf("failed to fetch users by IDs: %v", err)
        }
		return users, nil
	}

}

// SuggestUsers provides autocomplete suggestions
func (s *UserService) SuggestUsers(text string) (interface{}, error) {
	ctx := context.Background()
	return s.usersSearchService.Suggest(ctx, text)
}