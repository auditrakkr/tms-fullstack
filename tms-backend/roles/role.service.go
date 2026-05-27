package roles

import (
	"context"
	"fmt"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type RoleService struct {
	roleRepo repositories.Repository[models.Role]
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		roleRepo: repositories.NewRepository[models.Role](db),
	}
}

/* CREATE */

func (s *RoleService) CreateRole(ctx context.Context, createRoleDto *dto.CreateRoleDto) (*models.Role, error) {
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

func (s *RoleService) Update(ctx context.Context, roleId uint, updateRoleDto *dto.UpdateRoleDto) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}
	if err := copier.Copy(role, updateRoleDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	if err := s.roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("failed to update role: %v", err)
	}
	return role, nil
}

func (s *RoleService) Save(ctx context.Context, role *models.Role) (*models.Role, error) {
	role, err := s.roleRepo.Save(role)
	if err != nil {
		return nil, fmt.Errorf("failed to save role: %v", err)
	}
	return role, nil
}

/* DELETE */

func (s *RoleService) Delete(ctx context.Context, roleId uint) error {
	if err := s.roleRepo.Delete(roleId); err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
	}
	return nil
}

func (s *RoleService) Remove(ctx context.Context, role *models.Role) (*models.Role, error) {
	removed, err := s.roleRepo.Remove(role)
	if err != nil {
		return nil, fmt.Errorf("failed to remove role: %v", err)
	}
	return &removed, nil
}

/* READ */

func (s *RoleService) FindAllWithOptions(ctx context.Context, findOptions map[string]any) ([]models.Role, int64, error) {
	roles, total, err := s.roleRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get roles: %w", err)
	}
	return roles, total, nil
}

func (s *RoleService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.roleRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return roles, nil
}

func (s *RoleService) FindOne(ctx context.Context, roleId uint) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}
	return role, nil
}
