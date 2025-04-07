package regions

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)

type RegionService struct {
	regionRepo repositories.Repository[models.Region]
}

func NewRegionService() *RegionService {
	return &RegionService{
		regionRepo: repositories.Repository[models.Region]{DB: database.DB},
	}
}

/* CREATE */

func (s *RegionService) Create(createRegionDto *dto.CreateRegionDto) (*models.Region, error) {
	region := &models.Region{}
	if err := copier.Copy(region, createRegionDto); err != nil {
		return nil, err
	}
	region, err := s.regionRepo.Create(region)
	if err != nil {
		return nil, err
	}
	return region, nil

}

func (s *RegionService) InsertRegions(regions *[]dto.CreateRegionDto) (*[]models.Region, error) {
    var regionModels []models.Region

    // Copy data from DTOs to models
    if err := copier.Copy(&regionModels, regions); err != nil {
        return nil, err
    }

    // Perform bulk insert using GORM
    if err := s.regionRepo.DB.Create(&regionModels).Error; err != nil {
        // Handle other errors
        return nil, fmt.Errorf("there was a problem with Region(s) insertion: %w", err)
    }

    // Optionally clear any cache related to regions (if applicable)
    // Example: ClearCache("regions", "tenant-assignable-regions-info")

    return &regionModels, nil
}


/* UPDATE */

func (s *RegionService) Update(regionId uint, updateRegionDto *dto.UpdateRegionDto) (*models.Region, error) {
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %v", err)
	}
	if err := copier.Copy(region, updateRegionDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}
	err = s.regionRepo.Update(region)
	if err != nil {
		return nil, err
	}
	return region, nil
}

func (s *RegionService) Save (region *models.Region) (*models.Region, error) {
	region, err := s.regionRepo.Save(region)
	if err != nil {
		return nil, fmt.Errorf("failed to save region: %v", err)
	}
	return region, nil
}

/* READ */

func (s *RegionService) FindAllWithOptions(findOptions map[string]any) ([]models.Region, int64, error) {
	regions, totalCount, err := s.regionRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all regions with options: %w", err)
	}
	return regions, totalCount, nil
}

func (s *RegionService) GetAllRegions() ([]models.Region, error) {
	regions, err := s.regionRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all regions: %w", err)
	}
	return regions, nil
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


/* DELETE */
func (s *RegionService) Delete(regionId uint) error {
	err := s.regionRepo.Delete(regionId)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}
	return nil
}