package billings

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)


type BillingService struct {
	billRepo repositories.Repository[models.Billing]
}

func NewBillingService() *BillingService {
	return &BillingService{
		billRepo: repositories.Repository[models.Billing]{DB: database.DB},
	}
}

func (s *BillingService) CreateBilling(createBillingDto *dto.CreateBillingDto) (*models.Billing, error) {
	newBilling := &models.Billing{}
	if err := copier.Copy(newBilling, createBillingDto); err != nil {
		return nil, err
	}
	newBilling, err := s.billRepo.Create(newBilling)
	if err != nil {
		return nil, err
	}
	return newBilling, nil
}