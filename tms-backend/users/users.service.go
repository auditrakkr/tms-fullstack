package users

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)


type UserService struct {
	userRepo repositories.Repository[models.User]
	roleRepo repositories.Repository[models.Role]
	tenantRepo repositories.Repository[models.Tenant]
	tenantTeamRepo repositories.Repository[models.TenantTeam]
	tenantAccountOfficerRepo repositories.Repository[models.TenantAccountOfficer]
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repositories.Repository[models.User]{DB: database.DB},
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
	return newUser, nil
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