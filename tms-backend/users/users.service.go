package users

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
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

func (s *UserService) CreateUser(createUserDto *dto.CreateUserDto) (*models.User, error) {
	newUser := &models.User{
		FirstName: createUserDto.FirstName,
		LastName: createUserDto.LastName,
		PrimaryEmailAddress: createUserDto.PrimaryEmailAddress,
	}
}