package tenantconfigdetails

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)


type TenantConfigDetailsService struct {
	tenantConfigDetailsRepo repositories.Repository[models.TenantConfigDetail]

}

func NewTenantConfigDetailsService() *TenantConfigDetailsService {
	return &TenantConfigDetailsService{
		tenantConfigDetailsRepo: repositories.Repository[models.TenantConfigDetail]{DB: database.DB},
	}
}

/* CREATE */

func (s *TenantConfigDetailsService) CreateTenantConfigDetail(createTenantConfigDetailDto *dto.CreateTenantConfigDetailDto) (*models.TenantConfigDetail, error) {
	newTenantConfigDetail := &models.TenantConfigDetail{}
	if err := copier.Copy(newTenantConfigDetail, createTenantConfigDetailDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	newTenantConfigDetail, err := s.tenantConfigDetailsRepo.Create(newTenantConfigDetail)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant config detail: %v", err)
	}
	return newTenantConfigDetail, nil
}

/* UPDATE */

func (s *TenantConfigDetailsService) Update(id uint, tenantConfigDetail *dto.CreateTenantConfigDetailDto) (*models.TenantConfigDetail, error) {
	tenantConfigDetailToUpdate, err := s.tenantConfigDetailsRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant config detail: %v", err)
	}
	if err := copier.Copy(tenantConfigDetailToUpdate, tenantConfigDetail); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	err = s.tenantConfigDetailsRepo.Update(tenantConfigDetailToUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant config detail: %v", err)
	}
	return tenantConfigDetailToUpdate, nil
}

func (s *TenantConfigDetailsService) Save (tenantConfigDetail *models.TenantConfigDetail) (*models.TenantConfigDetail, error) {
	tenantConfigDetail, err := s.tenantConfigDetailsRepo.Save(tenantConfigDetail)
	if err != nil {
		return nil, fmt.Errorf("failed to save tenant config detail: %v", err)
	}
	return tenantConfigDetail, nil
}

/* READ */

func (s *TenantConfigDetailsService) FindAllWithOptions(findOptions map[string]any) ([]models.TenantConfigDetail, int64, error) {
	tenantConfigDetails, totalCount, err := s.tenantConfigDetailsRepo.FindAndCount(findOptions)
	if err != nil {
		return nil,0,fmt.Errorf("failed to find tenant config details: %v", err)
	}
	return tenantConfigDetails,totalCount,  nil
}

func (s *TenantConfigDetailsService) GetAllTenantConfigDetails() ([]models.TenantConfigDetail, error) {
	tenantConfigDetails, err := s.tenantConfigDetailsRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenant config details: %w", err)
	}
	return tenantConfigDetails, nil
}

func (s *TenantConfigDetailsService) FindOne(id uint) (*models.TenantConfigDetail, error) {
	tenantConfigDetail, err := s.tenantConfigDetailsRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant config detail: %w", err)
	}
	if tenantConfigDetail == nil {
		return nil, fmt.Errorf("tenant config detail not found")
	}
	return tenantConfigDetail, nil
}


/* DELETE */
func (s *TenantConfigDetailsService) Delete(id uint) error {
	err := s.tenantConfigDetailsRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant config detail: %w", err)
	}
	return nil
}