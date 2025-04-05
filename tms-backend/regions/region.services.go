package regions

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
)

type RegionService struct {
	regionRepo repositories.Repository[models.Region]
}

func NewRegionService() *RegionService {
	return &RegionService{
		regionRepo: repositories.Repository[models.Region]{DB: database.DB},
	}
}

func (s *RegionService) FindByRegionName(regionName string) (*models.Region, error) {
	/* region, err := s.regionRepo.FindByID(regionName)
	if err != nil {
		return nil, err
	}
	return region, nil */
	/* var region models.Region
	err := s.regionRepo.DB.Where("name = ?", regionName).First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil */
	var region models.Region
	err := s.regionRepo.CreateQueryBuilder().
		Where("name = ?", regionName).Find(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}