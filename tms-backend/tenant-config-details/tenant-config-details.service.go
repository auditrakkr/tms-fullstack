package tenantconfigdetails

import (
	"context"
	"fmt"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type TenantConfigDetailsService struct {
	tenantConfigDetailsRepo repositories.Repository[models.TenantConfigDetail]
}

func NewTenantConfigDetailsService(db *gorm.DB) *TenantConfigDetailsService {
	return &TenantConfigDetailsService{
		tenantConfigDetailsRepo: repositories.NewRepository[models.TenantConfigDetail](db),
	}
}

/* CREATE */

func (s *TenantConfigDetailsService) CreateTenantConfigDetail(ctx context.Context, createDto *dto.CreateTenantConfigDetailDto) (*models.TenantConfigDetail, error) {
	newDetail := &models.TenantConfigDetail{}
	if err := copier.Copy(newDetail, createDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	newDetail, err := s.tenantConfigDetailsRepo.Create(newDetail)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant config detail: %w", err)
	}
	return newDetail, nil
}

/* UPDATE */

func (s *TenantConfigDetailsService) Update(ctx context.Context, id uint, updateDto *dto.CreateTenantConfigDetailDto) (*models.TenantConfigDetail, error) {
	detail, err := s.tenantConfigDetailsRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant config detail: %w", err)
	}
	if err := copier.Copy(detail, updateDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	if err := s.tenantConfigDetailsRepo.Update(detail); err != nil {
		return nil, fmt.Errorf("failed to update tenant config detail: %w", err)
	}
	return detail, nil
}

func (s *TenantConfigDetailsService) Save(ctx context.Context, detail *models.TenantConfigDetail) (*models.TenantConfigDetail, error) {
	detail, err := s.tenantConfigDetailsRepo.Save(detail)
	if err != nil {
		return nil, fmt.Errorf("failed to save tenant config detail: %w", err)
	}
	return detail, nil
}

/* READ */

func (s *TenantConfigDetailsService) FindAllWithOptions(ctx context.Context, findOptions map[string]any) ([]models.TenantConfigDetail, int64, error) {
	details, total, err := s.tenantConfigDetailsRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find tenant config details: %w", err)
	}
	return details, total, nil
}

func (s *TenantConfigDetailsService) GetAllTenantConfigDetails(ctx context.Context) ([]models.TenantConfigDetail, error) {
	details, err := s.tenantConfigDetailsRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config details: %w", err)
	}
	return details, nil
}

func (s *TenantConfigDetailsService) FindOne(ctx context.Context, id uint) (*models.TenantConfigDetail, error) {
	detail, err := s.tenantConfigDetailsRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant config detail: %w", err)
	}
	return detail, nil
}

/* DELETE */

func (s *TenantConfigDetailsService) Delete(ctx context.Context, id uint) error {
	if err := s.tenantConfigDetailsRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete tenant config detail: %w", err)
	}
	return nil
}
