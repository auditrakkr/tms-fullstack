package roles

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)

type RoleService struct {
	roleRepo repositories.Repository[models.Role]
}

func NewRoleService() *RoleService {
	return &RoleService{
		roleRepo: repositories.Repository[models.Role]{DB: database.DB},
	}
}

/* CREATE */
func (s *RoleService) CreateRole(createRoleDto *dto.CreateUserDto) (*models.Role, error) {
	newRole := &models.Role{}
	if err := copier.Copy(newRole, createRoleDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}

	newRole, err := s.roleRepo.Create(newRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %v", err)
	}
	return newRole, nil
}

/* UPDATE */
func (s *RoleService) Update(roleId uint, updateRoleDto *dto.UpdateUserDto) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}

	if err := copier.Copy(role, updateRoleDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}

	err = s.roleRepo.Update(role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %v", err)
	}
	return role, nil
}

func (s *RoleService) Save (role *models.Role) (*models.Role, error) {
	role, err := s.roleRepo.Save(role)
	if err != nil {
		return nil, fmt.Errorf("failed to save role: %v", err)
	}
	return role, nil
}

/* DELETE */

func (s *RoleService) Delete(roleId uint) error {
	err := s.roleRepo.Delete(roleId)
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
	}
	return nil
}

func (s *RoleService) Remove (role *models.Role) (*models.Role, error) {
	removedRole, err := s.roleRepo.Remove(role)
	if err != nil {
		return nil, fmt.Errorf("failed to remove role: %v", err)
	}
	return &removedRole, nil
}

/* READ */
func (s *RoleService) FindAllWithOptions(findOptions map[string]any) ([]models.Role, int64, error) {
	roles, totalCount, err := s.roleRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all roles with options: %w", err)
	}
	return roles, totalCount, nil
}

func (s *RoleService) GetAllRoles() ([]models.Role, error) {
	roles, err := s.roleRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	return roles, nil
}

func (s *RoleService) FindOne (roleId uint) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}




