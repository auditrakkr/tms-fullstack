package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ResetPasswordResult is returned by ResetPassword so the controller can
// render the appropriate HTML view without the service knowing about HTTP.
type ResetPasswordResult struct {
	TokenExpired     bool
	NewPasswordSaved bool
	FormReturnURL    string // non-empty when the reset form should be shown
}

// UserService handles all user-domain business logic.
type UserService struct {
	userRepo                 repositories.Repository[models.User]
	roleRepo                 repositories.Repository[models.Role]
	tenantRepo               repositories.Repository[models.Tenant]
	tenantTeamRepo           repositories.Repository[models.TenantTeam]
	tenantAccountOfficerRepo repositories.Repository[models.TenantAccountOfficer]
	usersSearchService       *search.UsersSearchService
	db                       *gorm.DB
}

// NewUserService wires dependencies. esNode is the Elasticsearch node URL;
// pass an empty string to disable search indexing gracefully.
func NewUserService(db *gorm.DB, esNode string) *UserService {
	svc := &UserService{
		userRepo:                 repositories.NewRepository[models.User](db),
		roleRepo:                 repositories.NewRepository[models.Role](db),
		tenantRepo:               repositories.NewRepository[models.Tenant](db),
		tenantTeamRepo:           repositories.NewRepository[models.TenantTeam](db),
		tenantAccountOfficerRepo: repositories.NewRepository[models.TenantAccountOfficer](db),
		db:                       db,
	}

	searchService, err := search.NewUsersSearchService(esNode)
	if err != nil {
		slog.Warn("elasticsearch unavailable — search disabled", "error", err)
	} else {
		if err := searchService.Initialize(context.Background()); err != nil {
			slog.Warn("failed to initialise elasticsearch index", "error", err)
		} else {
			svc.usersSearchService = searchService
		}
	}
	return svc
}

// NewUserServiceWithDB is a convenience constructor for tests/migration helpers
// that already have database.DB available.
func NewUserServiceFromGlobal() *UserService {
	return NewUserService(database.DB, "")
}

/* ── CREATE ─────────────────────────────────────────────── */

func (s *UserService) CreateUser(ctx context.Context, createUserDto *dto.CreateUserDto, baseURL string) (*models.User, error) {
	newUser := &models.User{}
	if err := copier.Copy(newUser, createUserDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}

	// Hash the password before persisting.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	newUser.PasswordHash = string(hashedPassword)

	user, err := s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	invalidateUserCache()

	if s.usersSearchService != nil {
		if err := s.usersSearchService.IndexUser(ctx, *user); err != nil {
			slog.Warn("failed to index new user in elasticsearch", "userID", user.ID, "error", err)
		}
	}

	if global.AUTO_SEND_CONFIRM_EMAIL {
		if _, err := s.ConfirmEmailRequest(ctx, &createUserDto.PrimaryEmailAddress, 0, true, baseURL); err != nil {
			slog.Warn("failed to send confirmation email", "email", createUserDto.PrimaryEmailAddress, "error", err)
		}
	}

	return user, nil
}

/* ── UPDATE ─────────────────────────────────────────────── */

func (s *UserService) Update(ctx context.Context, userId uint, updateUserDto *dto.UpdateUserDto) (*models.User, error) {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if err := copier.Copy(user, updateUserDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	invalidateUserCache()
	if s.usersSearchService != nil {
		if err := s.usersSearchService.Update(ctx, *user); err != nil {
			slog.Warn("failed to update user in elasticsearch", "userID", user.ID, "error", err)
		}
	}
	return user, nil
}

/* ── DELETE ─────────────────────────────────────────────── */

func (s *UserService) DeleteUser(ctx context.Context, userId uint) error {
	if s.usersSearchService != nil {
		if err := s.usersSearchService.Remove(ctx, int(userId)); err != nil {
			slog.Warn("failed to remove user from elasticsearch", "userID", userId, "error", err)
		}
	}
	if err := s.userRepo.Delete(userId); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	invalidateUserCache()
	return nil
}

func (s *UserService) RemoveUser(ctx context.Context, user *models.User) (*models.User, error) {
	removed, err := s.userRepo.Remove(user)
	if err != nil {
		return nil, fmt.Errorf("failed to remove user: %w", err)
	}
	return &removed, nil
}

/* ── READ ────────────────────────────────────────────────── */

func (s *UserService) FindAllWithOptions(ctx context.Context, findOptions map[string]any) ([]models.User, int64, error) {
	users, total, err := s.userRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find users: %w", err)
	}
	return users, total, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	// Check Redis cache first.
	if database.Cache != nil {
		var cached []models.User
		if ok, _ := database.Cache.Get("users", &cached); ok {
			return cached, nil
		}
	}
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	if database.Cache != nil {
		database.Cache.Set("users", users, 25000) // 25 s
	}
	return users, nil
}

func (s *UserService) FindOne(ctx context.Context, userId uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

/* ── ROLES ───────────────────────────────────────────────── */

func (s *UserService) CreateAndAddRole(ctx context.Context, userId uint, createRoleDto *dto.CreateRoleDto) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newRole := &models.Role{}
	if err := copier.Copy(newRole, createRoleDto); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to map dto: %w", err)
	}
	if err := tx.Create(newRole).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role with this name already exists: %w", err)
		}
		return fmt.Errorf("failed to create role: %w", err)
	}
	return tx.Commit().Error
}

func (s *UserService) AddRoleById(ctx context.Context, userId, roleId uint) error {
	role := &models.Role{}
	if err := s.db.First(role, roleId).Error; err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}
	err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("Roles").Append(role)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("role already assigned to user: %w", err)
		}
		return fmt.Errorf("failed to add role to user: %w", err)
	}
	return nil
}

func (s *UserService) AddRolesById(ctx context.Context, userId uint, roleIds []uint) ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Where("id IN ?", roleIds).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to find roles: %w", err)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found")
	}
	if err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("Roles").Append(roles); err != nil {
		return nil, fmt.Errorf("failed to add roles: %w", err)
	}
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

func (s *UserService) RemoveRoleById(ctx context.Context, userId, roleId uint) ([]models.Role, error) {
	role := &models.Role{}
	if err := s.db.First(role, roleId).Error; err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	if err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("Roles").Delete(role); err != nil {
		return nil, fmt.Errorf("failed to remove role: %w", err)
	}
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

func (s *UserService) RemoveRolesById(ctx context.Context, userId uint, roleIds []uint) ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Where("id IN ?", roleIds).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to find roles: %w", err)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found")
	}
	if err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("Roles").Delete(roles); err != nil {
		return nil, fmt.Errorf("failed to remove roles: %w", err)
	}
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

/* ── LOOKUP HELPERS ─────────────────────────────────────── */

// FindByPrimaryEmailAddress satisfies the auth.UserLookup interface.
func (s *UserService) FindByPrimaryEmailAddress(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("primary_email_address = ?", email).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (s *UserService) FindByConfirmedPrimaryEmailAddress(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("primary_email_address = ? AND is_primary_email_verified = true", email).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find confirmed user: %w", err)
	}
	return &user, nil
}

func (s *UserService) FindByResetPassToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("reset_password_token = ?", token).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("user not found")
		}
		return nil, fmt.Errorf("failed to find user by reset token: %w", err)
	}
	return &user, nil
}

func (s *UserService) FindByPrimaryEmailVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("primary_email_verification_token = ?", token).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (s *UserService) FindByBackupEmailVerificationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("backup_email_verification_token = ?", token).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (s *UserService) FindById(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Select("*").
		Preload("Roles").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// FindByGoogleID satisfies the auth.UserLookup interface.
func (s *UserService) FindByGoogleID(ctx context.Context, googleId string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Select("*").
		Preload("Roles").
		Preload("GoogleProfile").
		Joins("JOIN google_profiles ON google_profiles.user_id = users.id").
		Where("google_profiles.google_id = ?", googleId).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("user not found")
		}
		return nil, fmt.Errorf("failed to get user by google ID: %w", err)
	}
	return &user, nil
}

// FindByFacebookID satisfies the auth.UserLookup interface.
func (s *UserService) FindByFacebookID(ctx context.Context, facebookId string) (*models.User, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Select("*").
		Preload("Roles").
		Preload("FacebookProfile").
		Joins("JOIN facebook_profiles ON facebook_profiles.user_id = users.id").
		Where("facebook_profiles.facebook_id = ?", facebookId).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("user not found")
		}
		return nil, fmt.Errorf("failed to get user by facebook ID: %w", err)
	}
	return &user, nil
}

/* ── TOKEN / CREDENTIAL HELPERS ─────────────────────────── */

// SetRefreshTokenHash satisfies the auth.UserLookup interface.
func (s *UserService) SetRefreshTokenHash(ctx context.Context, userId uint, refreshToken string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash refresh token: %w", err)
	}
	return s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Update("refresh_token_hash", string(hashed)).Error
}

func (s *UserService) SetGoogleProfile(ctx context.Context, userId uint, googleProfile *auth_dto.GoogleProfileDto) (*models.User, error) {
	newGP := &models.GoogleProfile{}
	if err := copier.Copy(newGP, googleProfile); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	newGP.UserID = userId
	if err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("GoogleProfile").Replace(newGP); err != nil {
		return nil, fmt.Errorf("failed to set google profile: %w", err)
	}
	return s.userRepo.FindByID(userId)
}

func (s *UserService) SetFacebookProfile(ctx context.Context, userId uint, fbProfile *auth_dto.FacebookProfileDto) (*models.User, error) {
	newFP := &models.FacebookProfile{}
	if err := copier.Copy(newFP, fbProfile); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	newFP.UserID = userId
	if err := s.userRepo.CreateQueryBuilder().Where("id = ?", userId).Association("FacebookProfile").Replace(newFP); err != nil {
		return nil, fmt.Errorf("failed to set facebook profile: %w", err)
	}
	return s.userRepo.FindByID(userId)
}

func (s *UserService) SetUserPassword(ctx context.Context, userId uint, password string) (bool, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}
	err = s.userRepo.CreateQueryBuilder().
		Where("id = ?", userId).
		Update("password_hash", string(hashed)).Error
	return err == nil, err
}

/* ── PHOTO ───────────────────────────────────────────────── */

func (s *UserService) SetUserPhoto(ctx context.Context, userId uint, file *multipart.FileHeader) error {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if file.Size > global.PHOTO_FILE_SIZE_LIMIT {
		return global.BadRequest(fmt.Sprintf("file too large (max %d bytes)", global.PHOTO_FILE_SIZE_LIMIT))
	}

	fileName := file.Filename
	if user.Photo != "" {
		fileName = user.Photo
	} else {
		fileName = fmt.Sprintf("%d_%s", user.ID, file.Filename)
	}

	uploadDir := filepath.Join(global.UPLOAD_DIRECTORY, "photos")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dest := filepath.Join(uploadDir, fileName)
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	user.Photo = fileName
	user.PhotoMimeType = file.Header.Get("Content-Type")
	return s.userRepo.Update(user)
}

// GetUserPhoto returns the path and MIME type for the user's photo.
// If no photo exists, it returns the default avatar path with an empty MIME type.
func (s *UserService) GetUserPhoto(ctx context.Context, userId uint) (filePath, mimeType string, err error) {
	user, err := s.userRepo.FindByID(userId)
	if err != nil {
		return "", "", fmt.Errorf("failed to find user: %w", err)
	}
	defaultPath := filepath.Join(global.UPLOAD_DIRECTORY, "photos", "blankPhotoAvatar.png")
	if user.Photo == "" {
		return defaultPath, "", nil
	}
	photoPath := filepath.Join(global.UPLOAD_DIRECTORY, "photos", user.Photo)
	if _, statErr := os.Stat(photoPath); os.IsNotExist(statErr) {
		return defaultPath, "", nil
	}
	return photoPath, user.PhotoMimeType, nil
}

/* ── PASSWORD RESET ──────────────────────────────────────── */

// ResetPasswordRequest sends the reset link email and returns a safe notification.
// baseURL is the full scheme+host (e.g. "https://example.com") built by the controller.
func (s *UserService) ResetPasswordRequest(ctx context.Context, email, baseURL string) (*global.GenericNotificationResponse, error) {
	safeResponse := &global.GenericNotificationResponse{
		NotificationClass:   "is-success",
		NotificationMessage: fmt.Sprintf("If your email %s is found, you will receive a password reset link shortly", email),
	}

	user, err := s.FindByPrimaryEmailAddress(ctx, email)
	if err != nil {
		// Don't reveal whether the email exists.
		return safeResponse, nil
	}

	token, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	user.ResetPasswordToken = token
	user.ResetPasswordExpiration = time.Now().Add(global.PASSWORD_RESET_EXPIRATION)
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to save reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s%s/users/reset-password/%s", baseURL, global.API_BASE_URL, token)
	mailText := strings.Replace(global.ResetPasswordMailOptionSettings.TextTemplate, "{url}", resetURL, -1)
	global.SendMailAsync(global.MailOptions{
		To:      user.PrimaryEmailAddress,
		From:    global.ResetPasswordMailOptionSettings.From,
		Subject: global.ResetPasswordMailOptionSettings.Subject,
		Text:    mailText,
	})

	return safeResponse, nil
}

// ResetPassword validates the token and optionally sets a new password.
// The returned ResetPasswordResult carries all the data the controller needs
// to render the appropriate HTML view — the service itself does not write HTTP.
func (s *UserService) ResetPassword(ctx context.Context, token string, newPassword *string, baseURL string) (*ResetPasswordResult, error) {
	var user models.User
	err := s.userRepo.CreateQueryBuilder().
		Where("reset_password_token = ?", token).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("invalid token")
		}
		return nil, global.Internal(err)
	}

	if user.ResetPasswordExpiration.Before(time.Now()) {
		return &ResetPasswordResult{TokenExpired: true}, nil
	}

	if newPassword != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, global.Internal(err)
		}
		user.PasswordHash = string(hashed)
		user.ResetPasswordToken = ""
		user.ResetPasswordExpiration = time.Time{}
		if err := s.userRepo.Update(&user); err != nil {
			return nil, global.Internal(err)
		}
		return &ResetPasswordResult{NewPasswordSaved: true}, nil
	}

	// Show the reset form.
	returnURL := fmt.Sprintf("%s%s/users/reset-password/%s", baseURL, global.API_BASE_URL, token)
	return &ResetPasswordResult{FormReturnURL: returnURL}, nil
}

/* ── EMAIL VERIFICATION ──────────────────────────────────── */

// ConfirmEmailRequest sends a verification link and returns a safe notification.
func (s *UserService) ConfirmEmailRequest(ctx context.Context, email *string, userId uint, isPrimary bool, baseURL string) (*global.GenericNotificationResponse, error) {
	safeResponse := &global.GenericNotificationResponse{
		NotificationClass:   "is-info",
		NotificationMessage: "If a valid user is found, a verification email will be sent shortly",
	}

	var (
		user *models.User
		err  error
	)
	switch {
	case userId > 0:
		user, err = s.userRepo.FindByID(userId)
	case email != nil && isPrimary:
		user, err = s.FindByPrimaryEmailAddress(ctx, *email)
	case email != nil:
		var found models.User
		dbErr := s.userRepo.CreateQueryBuilder().
			Where("backup_email_address = ?", *email).First(&found).Error
		if dbErr == nil {
			user = &found
		}
		err = dbErr
	default:
		return nil, global.BadRequest("email or userId must be provided")
	}

	if err != nil || user == nil {
		return safeResponse, nil
	}

	token, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	if isPrimary {
		user.PrimaryEmailVerificationToken = token
	} else {
		user.BackupEmailVerificationToken = token
	}
	user.EmailVerificationTokenExpiration = time.Now().Add(global.EMAIL_VERIFICATION_EXPIRATION)
	if _, err := s.userRepo.Save(user); err != nil {
		return nil, fmt.Errorf("failed to save verification token: %w", err)
	}

	endpoint := "confirm-primary-email"
	if !isPrimary {
		endpoint = "confirm-backup-email"
	}

	verificationURL := fmt.Sprintf("%s%s/users/%s/%s", baseURL, global.API_BASE_URL, endpoint, token)
	mailText := strings.Replace(global.ConfirmEmailMailOptionSettings.TextTemplate, "{url}", verificationURL, -1)

	recipient := user.PrimaryEmailAddress
	if !isPrimary {
		recipient = user.BackupEmailAddress
	}
	global.SendMailAsync(global.MailOptions{
		To:      recipient,
		From:    global.ConfirmEmailMailOptionSettings.From,
		Subject: global.ConfirmEmailMailOptionSettings.Subject,
		Text:    mailText,
	})

	return safeResponse, nil
}

// ConfirmEmailResult carries the data the controller needs to render the
// confirm-email-feedback.html view without the service knowing about HTTP.
type ConfirmEmailResult struct {
	Success bool
	Message string
}

// ConfirmEmail marks an email address as verified and returns a ConfirmEmailResult
// with a success flag and message that the controller uses to render
// users/confirm-email-feedback.html.
func (s *UserService) ConfirmEmail(ctx context.Context, token string, isPrimary bool) (*ConfirmEmailResult, error) {
	var user models.User
	var err error
	if isPrimary {
		err = s.userRepo.CreateQueryBuilder().
			Where("primary_email_verification_token = ?", token).First(&user).Error
	} else {
		err = s.userRepo.CreateQueryBuilder().
			Where("backup_email_verification_token = ?", token).First(&user).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ConfirmEmailResult{Success: false, Message: "Problem confirming email"}, nil
		}
		return &ConfirmEmailResult{Success: false, Message: "Problem confirming email"}, nil
	}
	if user.EmailVerificationTokenExpiration.Before(time.Now()) {
		return &ConfirmEmailResult{Success: false, Message: "Problem confirming email. Token has expired!"}, nil
	}
	if isPrimary {
		user.IsPrimaryEmailVerified = true
		user.PrimaryEmailVerificationToken = ""
	} else {
		user.IsBackupEmailVerified = true
		user.BackupEmailVerificationToken = ""
	}
	user.EmailVerificationTokenExpiration = time.Time{}
	if err := s.userRepo.Update(&user); err != nil {
		return &ConfirmEmailResult{Success: false, Message: "Problem confirming email!"}, nil
	}
	return &ConfirmEmailResult{Success: true, Message: "Email confirmed!"}, nil
}

/* ── OAUTH UPSERT ────────────────────────────────────────── */

// FindOrCreateGoogleUser satisfies the auth.UserLookup interface.
// It finds an existing user by Google ID or email, links the Google profile,
// and creates a new user account if none exists.
func (s *UserService) FindOrCreateGoogleUser(ctx context.Context, profile *auth_dto.GoogleProfileDto) (*models.User, error) {
	// 1. Try to find by Google ID.
	if profile.GoogleID != nil {
		user, err := s.FindByGoogleID(ctx, *profile.GoogleID)
		if err == nil {
			return user, nil
		}
	}

	// 2. Try to find by email and link the profile.
	if profile.Email != nil {
		user, err := s.FindByPrimaryEmailAddress(ctx, *profile.Email)
		if err == nil {
			_, _ = s.SetGoogleProfile(ctx, user.ID, profile)
			return user, nil
		}
	}

	// 3. Create a new user.
	var firstName, lastName, email string
	if profile.GivenName != nil {
		firstName = *profile.GivenName
	}
	if profile.FamilyName != nil {
		lastName = *profile.FamilyName
	}
	if profile.Email != nil {
		email = *profile.Email
	}
	verified := profile.EmailVerified != nil && *profile.EmailVerified

	newUser := &models.User{
		FirstName:              firstName,
		LastName:               lastName,
		PrimaryEmailAddress:    email,
		IsPrimaryEmailVerified: verified,
		IsActive:               true,
	}
	user, err := s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from Google profile: %w", err)
	}
	_, _ = s.SetGoogleProfile(ctx, user.ID, profile)
	return user, nil
}

// FindOrCreateFacebookUser satisfies the auth.UserLookup interface.
func (s *UserService) FindOrCreateFacebookUser(ctx context.Context, profile *auth_dto.FacebookProfileDto) (*models.User, error) {
	// 1. Try to find by Facebook ID.
	if profile.FacebookID != nil {
		user, err := s.FindByFacebookID(ctx, *profile.FacebookID)
		if err == nil {
			return user, nil
		}
	}

	// 2. Try to find by email.
	if profile.Email != nil {
		user, err := s.FindByPrimaryEmailAddress(ctx, *profile.Email)
		if err == nil {
			_, _ = s.SetFacebookProfile(ctx, user.ID, profile)
			return user, nil
		}
	}

	// 3. Create a new user.
	var firstName, lastName, email string
	if profile.Name != nil {
		firstName = profile.Name.GivenName
		lastName = profile.Name.FamilyName
	}
	if profile.Email != nil {
		email = *profile.Email
	}

	newUser := &models.User{
		FirstName:           firstName,
		LastName:            lastName,
		PrimaryEmailAddress: email,
		IsActive:            true,
	}
	user, err := s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from Facebook profile: %w", err)
	}
	_, _ = s.SetFacebookProfile(ctx, user.ID, profile)
	return user, nil
}

/* ── SEARCH ──────────────────────────────────────────────── */

func (s *UserService) SearchForUsers(ctx context.Context, text string, returnRaw bool) (any, error) {
	if s.usersSearchService == nil {
		return []models.User{}, nil
	}
	results, err := s.usersSearchService.Search(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	if returnRaw {
		return results, nil
	}
	var ids []uint
	for _, r := range results {
		ids = append(ids, uint(r.ID))
	}
	if len(ids) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	if err := s.userRepo.CreateQueryBuilder().Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch users by IDs: %w", err)
	}
	return users, nil
}

func (s *UserService) SuggestUsers(ctx context.Context, text string) (any, error) {
	if s.usersSearchService == nil {
		return []any{}, nil
	}
	return s.usersSearchService.Suggest(ctx, text)
}

/* ── BULK INSERT ─────────────────────────────────────────── */

// InsertUsers creates multiple users in bulk, hashing passwords and sending
// confirmation emails for each.
func (s *UserService) InsertUsers(ctx context.Context, createUserDtos []dto.CreateUserDto, baseURL string) ([]models.User, error) {
	var usersToInsert []models.User
	for i := range createUserDtos {
		u := &models.User{}
		if err := copier.Copy(u, &createUserDtos[i]); err != nil {
			return nil, fmt.Errorf("failed to map dto at index %d: %w", i, err)
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password at index %d: %w", i, err)
		}
		u.PasswordHash = string(hashed)
		usersToInsert = append(usersToInsert, *u)
	}

	if err := s.db.Create(&usersToInsert).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, global.Conflict("one or more users conflict with existing records")
		}
		return nil, fmt.Errorf("failed to bulk insert users: %w", err)
	}

	// Index and send confirmation emails in the background.
	for i := range usersToInsert {
		u := usersToInsert[i]
		if s.usersSearchService != nil {
			if err := s.usersSearchService.IndexUser(ctx, u); err != nil {
				slog.Warn("failed to index user in elasticsearch", "userID", u.ID, "error", err)
			}
		}
		if global.AUTO_SEND_CONFIRM_EMAIL && u.PrimaryEmailAddress != "" {
			email := u.PrimaryEmailAddress
			if _, err := s.ConfirmEmailRequest(ctx, &email, 0, true, baseURL); err != nil {
				slog.Warn("failed to send confirmation email", "email", u.PrimaryEmailAddress, "error", err)
			}
		}
	}
	return usersToInsert, nil
}

/* ── USER-SIDE TENANT RELATIONS ──────────────────────────── */

// CreateAndSetTenantForPrimaryContact creates a new tenant with this user as primary contact.
func (s *UserService) CreateAndSetTenantForPrimaryContact(ctx context.Context, userID uint, createTenantDto *dto.CreateTenantDto) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	// Delegate to tenants package via a raw DB operation to avoid circular import.
	// The Tenant model is used directly.
	tenant := &models.Tenant{
		Name:             createTenantDto.Name,
		Address:          createTenantDto.Address,
		RegionName:       createTenantDto.RegionName,
		PrimaryContactID: user.ID,
	}
	if createTenantDto.MoreInfo != nil {
		tenant.MoreInfo = *createTenantDto.MoreInfo
	}
	if createTenantDto.Status != nil {
		tenant.Status = *createTenantDto.Status
	}
	return s.db.Create(tenant).Error
}

// SetAsPrimaryContactByTenantID sets this user as the primary contact for a tenant.
func (s *UserService) SetAsPrimaryContactByTenantID(ctx context.Context, userID, tenantID uint) ([]models.Tenant, error) {
	if err := s.db.Model(&models.Tenant{}).Where("id = ?", tenantID).
		Update("primary_contact_id", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to set primary contact: %w", err)
	}
	var tenants []models.Tenant
	if err := s.db.Where("primary_contact_id = ?", userID).Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user's primary contact tenants: %w", err)
	}
	return tenants, nil
}

// RemoveAsPrimaryContactByTenantID removes this user as primary contact from a tenant.
func (s *UserService) RemoveAsPrimaryContactByTenantID(ctx context.Context, userID, tenantID uint) ([]models.Tenant, error) {
	if err := s.db.Model(&models.Tenant{}).
		Where("id = ? AND primary_contact_id = ?", tenantID, userID).
		Update("primary_contact_id", nil).Error; err != nil {
		return nil, fmt.Errorf("failed to remove primary contact: %w", err)
	}
	var tenants []models.Tenant
	if err := s.db.Where("primary_contact_id = ?", userID).Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user's primary contact tenants: %w", err)
	}
	return tenants, nil
}

/* ── USER-SIDE TEAM MEMBERSHIPS ──────────────────────────── */

// CreateTenantAndSetTeamMembership creates a new tenant and adds this user as a team member.
func (s *UserService) CreateTenantAndSetTeamMembership(ctx context.Context, userID uint, createTenantTeamDto *dto.CreateTenantTeamDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if createTenantTeamDto.Tenant == nil {
			return global.BadRequest("tenant data is required")
		}
		tenant := &models.Tenant{
			Name:       createTenantTeamDto.Tenant.Name,
			RegionName: createTenantTeamDto.Tenant.RegionName,
		}
		if err := tx.Create(tenant).Error; err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		team := &models.TenantTeam{
			TenantID:         tenant.ID,
			UserID:           userID,
			Roles:            createTenantTeamDto.Roles,
			TenantUniqueName: createTenantTeamDto.TenantUniqueName,
			TenantUniqueID:   uint(createTenantTeamDto.TenantUniqueId),
		}
		return tx.Create(team).Error
	})
}

// SetTeamMembershipByID creates a TenantTeam record linking this user to a tenant.
func (s *UserService) SetTeamMembershipByID(ctx context.Context, userID, tenantID uint, roles []global.TenantTeamRole) error {
	team := &models.TenantTeam{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles,
	}
	return s.db.Create(team).Error
}

// DeleteTeamMembershipByID removes a user's team membership from a tenant.
func (s *UserService) DeleteTeamMembershipByID(ctx context.Context, userID, tenantID uint) ([]models.TenantTeam, error) {
	if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Delete(&models.TenantTeam{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete team membership: %w", err)
	}
	var memberships []models.TenantTeam
	if err := s.db.Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch team memberships: %w", err)
	}
	return memberships, nil
}

// UpdateTeamMembershipRolesByID updates the roles for a user's team membership.
func (s *UserService) UpdateTeamMembershipRolesByID(ctx context.Context, userID, tenantID uint, roles []global.TenantTeamRole) ([]models.TenantTeam, error) {
	if err := s.db.Model(&models.TenantTeam{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Update("roles", roles).Error; err != nil {
		return nil, fmt.Errorf("failed to update team membership roles: %w", err)
	}
	var memberships []models.TenantTeam
	if err := s.db.Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch team memberships: %w", err)
	}
	return memberships, nil
}

/* ── USER-SIDE ACCOUNT OFFICER ASSIGNMENTS ───────────────── */

// CreateAndSetAccountOfficerForWhichTenant creates a tenant and assigns this user as an account officer.
func (s *UserService) CreateAndSetAccountOfficerForWhichTenant(ctx context.Context, userID uint, dto *dto.CreateTenantAccountOfficerDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if dto.Tenant == nil {
			return global.BadRequest("tenant data is required")
		}
		tenant := &models.Tenant{Name: dto.Tenant.Name, RegionName: dto.Tenant.RegionName}
		if err := tx.Create(tenant).Error; err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		officer := &models.TenantAccountOfficer{
			TenantID: tenant.ID,
			UserID:   userID,
			Roles:    dto.Roles,
		}
		return tx.Create(officer).Error
	})
}

// SetTenantAccountOfficerForWhichTenantByID adds the user as an account officer for an existing tenant.
func (s *UserService) SetTenantAccountOfficerForWhichTenantByID(ctx context.Context, userID, tenantID uint, dto *dto.CreateTenantAccountOfficerDto) error {
	officer := &models.TenantAccountOfficer{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    dto.Roles,
	}
	return s.db.Create(officer).Error
}

// DeleteTenantAccountOfficerForWhichTenantByID removes the user's account officer assignment.
func (s *UserService) DeleteTenantAccountOfficerForWhichTenantByID(ctx context.Context, userID, tenantID uint) error {
	return s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Delete(&models.TenantAccountOfficer{}).Error
}

// UpdateTenantAccountOfficerForWhichTenantByID updates the account officer roles.
func (s *UserService) UpdateTenantAccountOfficerForWhichTenantByID(ctx context.Context, userID, tenantID uint, dto *dto.UpdateTenantAccountOfficerDto) error {
	return s.db.Model(&models.TenantAccountOfficer{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Update("roles", dto.Roles).Error
}

/* ── PRIVATE HELPERS ─────────────────────────────────────── */

func invalidateUserCache() {
	if database.Cache != nil {
		database.Cache.Delete("users") //nolint:errcheck
	}
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
