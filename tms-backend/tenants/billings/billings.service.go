package billings

import (
	"context"
	"fmt"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type BillingService struct {
	billRepo repositories.Repository[models.Billing]
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{
		billRepo: repositories.NewRepository[models.Billing](db),
	}
}

func (s *BillingService) CreateBilling(ctx context.Context, createBillingDto *dto.CreateBillingDto) (*models.Billing, error) {
	newBilling := &models.Billing{}
	if err := copier.Copy(newBilling, createBillingDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	newBilling, err := s.billRepo.Create(newBilling)
	if err != nil {
		return nil, fmt.Errorf("failed to create billing: %w", err)
	}
	return newBilling, nil
}

func (s *BillingService) FindAll(ctx context.Context) ([]models.Billing, error) {
	billings, err := s.billRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get billings: %w", err)
	}
	return billings, nil
}
