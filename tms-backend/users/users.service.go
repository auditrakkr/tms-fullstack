package users

import (
	"context"
	"fmt"

	auth_dto "github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/auditrakkr/tms-fullstack/tms-backend/search"

	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo                 repositories.Repository[models.User]
	roleRepo                 repositories.Repository[models.Role]
	tenantRepo               repositories.Repository[models.Tenant]
	tenantTeamRepo           repositories.Repository[models.TenantTeam]
	tenantAccountOfficerRepo repositories.Repository[models.TenantAccountOfficer]
	usersSearchService *search.UsersSearchService
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
		userRepo: repositories.Repository[models.User]{DB: database.DB},
		roleRepo: repositories.Repository[models.Role]{DB: database.DB},
		tenantRepo: repositories.Repository[models.Tenant]{DB: database.DB},
		tenantTeamRepo: repositories.Repository[models.TenantTeam]{DB: database.DB},
		tenantAccountOfficerRepo: repositories.Repository[models.TenantAccountOfficer]{DB: database.DB},
		usersSearchService: searchService,
	}
}

/* Create  */

func (s *UserService) CreateUser(createUserDto *dto.CreateUserDto) (*models.User, error) {
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
	return user, nil
}

/* Delete */

func (s *UserService) DeleteUser(userId uint) error {
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