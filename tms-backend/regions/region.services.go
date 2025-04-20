package regions

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type RegionService struct {
	regionRepo             repositories.Repository[models.Region]
	tenantConfigDetailRepo repositories.Repository[models.TenantConfigDetail]
	cache                  *database.RedisCache
}

func NewRegionService() *RegionService {
	return &RegionService{
		regionRepo:             repositories.Repository[models.Region]{DB: database.DB},
		tenantConfigDetailRepo: repositories.Repository[models.TenantConfigDetail]{DB: database.DB},
		cache:                  database.Cache,
	}
}

/* CREATE */

func (s *RegionService) Create(createRegionDto *dto.CreateRegionDto) (*models.Region, error) {
	// Create a new region instance
	region := &models.Region{}

	// Copy data from DTO to model
	if err := copier.Copy(region, createRegionDto); err != nil {
		return nil, fmt.Errorf("failed to map DTO to model: %v", err)
	}

	// Handle encryption of sensitive data
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt sensitive data: %v", err)
	}

	// Save the region
	region, err := s.regionRepo.Create(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create region: %v", err)
	}

	// Clear cache
	if s.cache != nil {
		s.clearRegionCache()
	}

	return region, nil

}

func (s *RegionService) InsertRegions(regions *[]dto.CreateRegionDto) (*[]models.Region, error) {
	var regionModels []models.Region

	// Copy data from DTOs to models
	if err := copier.Copy(&regionModels, regions); err != nil {
		return nil, err
	}

	// Encrypt sensitive data in each region
	for i := range regionModels {
		if err := s.encryptSensitiveData(&regionModels[i]); err != nil {
			return nil, fmt.Errorf("failed to encrypt sensitive data: %v", err)
		}
	}

	// Perform bulk insert using GORM
	if err := s.regionRepo.DB.Create(&regionModels).Error; err != nil {
		// Handle other errors
		return nil, fmt.Errorf("there was a problem with Region(s) insertion: %w", err)
	}

	// Optionally clear any cache related to regions (if applicable)
	// Example: ClearCache("regions", "tenant-assignable-regions-info")
	// Clear cache
	if s.cache != nil {
		s.clearRegionCache()
	}

	return &regionModels, nil
}

/* UPDATE */

func (s *RegionService) Update(regionId uint, updateRegionDto *dto.UpdateRegionDto) (*models.Region, error) {
	// Find the region
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %v", err)
	}

	// Copy data from DTO to model
	if err := copier.Copy(region, updateRegionDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}

	// Handle encryption of sensitive data
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt sensitive data: %v", err)
	}

	// Update the region
	err = s.regionRepo.Update(region)
	if err != nil {
		return nil, err
	}

	// Clear cache
	if s.cache != nil {
		s.clearRegionCache()
	}

	return region, nil
}

func (s *RegionService) Save(region *models.Region) (*models.Region, error) {
	// Handle encryption of sensitive data
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt sensitive data: %v", err)
	}

	// Save the region
	region, err := s.regionRepo.Save(region)
	if err != nil {
		return nil, fmt.Errorf("failed to save region: %v", err)
	}

	// Clear cache
	if s.cache != nil {
		s.clearRegionCache()
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
	// Try to get from cache first
	if s.cache != nil {
		var regions []models.Region
		found, err := s.cache.Get("regions", &regions)
		if err == nil && found {
			return regions, nil
		}
	}

	// If not in cache or cache error, get from DB
	regions, err := s.regionRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all regions: %w", err)
	}

	// Store in cache
	if s.cache != nil {
		s.cache.Set("regions", regions, 25000) // 25 seconds cache
	}

	return regions, nil
}

func (s *RegionService) FindOne(id uint) (*models.Region, error) {
	region, err := s.regionRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}
	if region == nil {
		return nil, fmt.Errorf("region not found")
	}
	return region, nil
}

func (s *RegionService) FindByRegionName(regionName string) (*models.Region, error) {
	var region models.Region
	err := s.regionRepo.CreateQueryBuilder().
		Where("name = ?", regionName).
		First(&region).Error


	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region not found")
		}
		return nil, fmt.Errorf("failed to find region by name: %w", err)
	}

	return &region, nil
}

// GetTenantAssignableRegionsInfo returns regions info for tenant assignment
func (s *RegionService) GetTenantAssignableRegionsInfo() ([]models.Region, error) {
	// Try to get from cache first
	if s.cache != nil {
		var regions []models.Region
		found, err := s.cache.Get("tenant-assignable-regions-info", &regions)
		if err == nil && found {
			return regions, nil
		}
	}

	var regions []models.Region
	err := s.regionRepo.CreateQueryBuilder().
		Select("id, name, root_domain_name, description, country, city, tenant_count_capacity").
		Find(&regions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant assignable regions info: %w", err)
	}

	// Calculate tenant count for each region
	for i := range regions {
		var count int64
		err := s.tenantConfigDetailRepo.CreateQueryBuilder().
			Where("region_id = ?", regions[i].ID).
			Count(&count).Error

		if err != nil {
			return nil, fmt.Errorf("failed to count tenants for region: %w", err)
		}

		// We can't directly map this count, but we could add it to a field if needed
		// For now, we'll leave it as is
	}

	// Store in cache
	if s.cache != nil {
		s.cache.Set("tenant-assignable-regions-info", regions, 25000) // 25 seconds cache
	}

	return regions, nil
}


/* DELETE */
func (s *RegionService) Delete(regionId uint) error {
	err := s.regionRepo.Delete(regionId)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}
	return nil
}

/* ASSOCIATION section */
func (s *RegionService) AddTenantConfigDetailById(regionId uint, tenantConfigDetailId uint) error {
	// Get the region
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}

	// Get the tenant config detail
	tenantConfigDetail, err := s.tenantConfigDetailRepo.FindByID(tenantConfigDetailId)
	if err != nil {
		return fmt.Errorf("failed to find tenant config detail: %w", err)
	}

	// Add association
	tenantConfigDetail.RegionID = regionId
	tenantConfigDetail.Region = *region

	err = s.tenantConfigDetailRepo.Update(tenantConfigDetail)
	if err != nil {
		return fmt.Errorf("failed to add tenant config detail to region: %w", err)
	}

	// Clear cache
	if s.cache != nil {
		s.clearRegionCache()
	}

	return nil
}


func (s *RegionService) AddTenantConfigDetailsById(regionId uint, tenantConfigDetailIds []uint) ([]models.TenantConfigDetail, error) {
	// Get the region to update
    region, err := s.regionRepo.FindByID(regionId)
    if err != nil {
        return nil, fmt.Errorf("failed to find region: %w", err)
    }

    // Get the tenant config details to add
    var tenantConfigDetailsToAdd []models.TenantConfigDetail
	if len(tenantConfigDetailIds) > 0 {
		err := s.tenantConfigDetailRepo.CreateQueryBuilder().Where("id IN ?", tenantConfigDetailIds).Find(&tenantConfigDetailsToAdd).Error

		if err != nil {
			return nil, fmt.Errorf("failed to find tenant config details: %w", err)
		}

		// Check if any tenant config details were found
        if len(tenantConfigDetailsToAdd) == 0 {
            return nil, fmt.Errorf("no tenant config details found for the provided IDs")
        }
	}

	// Apply the association in a transaction to ensure consistency
    tx := s.tenantConfigDetailRepo.DB.Begin()
    if tx.Error != nil {
        return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
    }

    // Defer transaction rollback in case of error
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

	// Update each tenant config detail to associate with the region
    for i := range tenantConfigDetailsToAdd {
        tenantConfigDetailsToAdd[i].RegionID = regionId
        tenantConfigDetailsToAdd[i].Region = *region

        err := tx.Save(&tenantConfigDetailsToAdd[i]).Error
        if err != nil {
            tx.Rollback()
            // Check for unique constraint violation
            if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
                return nil, fmt.Errorf("tenant config detail already assigned to region: %w", err)
            }
            return nil, fmt.Errorf("failed to add tenant config detail to region: %w", err)
        }
    }

	// Commit the transaction
    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    // Fetch the updated tenant config details for the region
    var tenantConfigDetails []models.TenantConfigDetail
    err = s.tenantConfigDetailRepo.CreateQueryBuilder().
        Where("region_id = ?", regionId).
        Find(&tenantConfigDetails).Error

    if err != nil {
        return nil, fmt.Errorf("failed to get tenant config details after update: %w", err)
    }

    // Clear cache
    if s.cache != nil {
        s.clearRegionCache()
    }

    return tenantConfigDetails, nil
}


func (s *RegionService) RemoveTenantConfigDetailById(regionId uint, tenantConfigDetailId uint) error {
	// Get the tenant config detail
    tenantConfigDetail, err := s.tenantConfigDetailRepo.FindByID(tenantConfigDetailId)
    if err != nil {
        return fmt.Errorf("failed to find tenant config detail: %w", err)
    }

	// Check if this tenant config detail belongs to the specified region
    if tenantConfigDetail.RegionID != regionId {
        return fmt.Errorf("tenant config detail is not assigned to this region")
    }

	 // Remove association by setting RegionID to null
    tenantConfigDetail.RegionID = 0
    // Clear the relationship without deleting the actual record
    tenantConfigDetail.Region = models.Region{}

	// Update the tenant config detail
    err = s.tenantConfigDetailRepo.Update(tenantConfigDetail)
    if err != nil {
        return fmt.Errorf("failed to remove tenant config detail from region: %w", err)
    }

	// Clear cache
    if s.cache != nil {
        s.clearRegionCache()
    }

    return nil
}

func (s *RegionService) RemoveTenantConfigDetailsById(regionId uint, tenantConfigDetailIds []uint) error {
	// Perform removal in a transaction to ensure consistency
    tx := s.tenantConfigDetailRepo.DB.Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to start transaction: %w", tx.Error)
    }

    // Defer transaction rollback in case of error
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Get the tenant config details to remove
    var tenantConfigDetails []models.TenantConfigDetail
    if err := tx.Where("id IN ? AND region_id = ?", tenantConfigDetailIds, regionId).Find(&tenantConfigDetails).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to find tenant config details: %w", err)
    }

    // Skip if no tenant config details were found
    if len(tenantConfigDetails) == 0 {
        tx.Rollback() // No need to continue with transaction
        return nil
    }

    // Remove associations by setting RegionID to 0
    for i := range tenantConfigDetails {
        tenantConfigDetails[i].RegionID = 0
        tenantConfigDetails[i].Region = models.Region{}

        if err := tx.Save(&tenantConfigDetails[i]).Error; err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to remove association for tenant config detail %d: %w", tenantConfigDetails[i].ID, err)
        }
    }

    // Commit the transaction
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    // Clear cache
    if s.cache != nil {
        s.clearRegionCache()
    }

    return nil
}




/* Helper methods */

// encryptSensitiveData encrypts sensitive data in a region
func (s *RegionService) encryptSensitiveData(region *models.Region) error {
	// Only encrypt if content exists and isn't already encrypted

	// WebServerProperties password
	if region.WebServerProperties != nil && region.WebServerProperties.Password != nil &&
		region.WebServerProperties.Password.Content != nil && region.WebServerProperties.Password.IV == nil {
		encrypted, err := utils.Encrypt(*region.WebServerProperties.Password.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt webserver password: %v", err)
		}
		region.WebServerProperties.Password = encrypted
	}

	// DBProperties password
	if region.DBProperties.Password != nil && region.DBProperties.Password.Content != nil &&
		region.DBProperties.Password.IV == nil {
		encrypted, err := utils.Encrypt(*region.DBProperties.Password.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt database password: %v", err)
		}
		region.DBProperties.Password = encrypted
	}

	// ElasticSearchProperties password
	if region.ElasticSearchProperties != nil && region.ElasticSearchProperties.Password != nil &&
		region.ElasticSearchProperties.Password.Content != nil && region.ElasticSearchProperties.Password.IV == nil {
		encrypted, err := utils.Encrypt(*region.ElasticSearchProperties.Password.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt elasticsearch password: %v", err)
		}
		region.ElasticSearchProperties.Password = encrypted
	}

	// RedisProperties password
	if region.RedisProperties != nil && region.RedisProperties.Password != nil &&
		region.RedisProperties.Password.Content != nil && region.RedisProperties.Password.IV == nil {
		encrypted, err := utils.Encrypt(*region.RedisProperties.Password.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt redis password: %v", err)
		}
		region.RedisProperties.Password = encrypted
	}

	// RootFileSystem password
	if region.RootFileSystem != nil && region.RootFileSystem.Password != nil &&
		region.RootFileSystem.Password.Content != nil && region.RootFileSystem.Password.IV == nil {
		encrypted, err := utils.Encrypt(*region.RootFileSystem.Password.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt root file system password: %v", err)
		}
		region.RootFileSystem.Password = encrypted
	}

	// SMTPAuth password
	if region.SMTPAuth != nil && region.SMTPAuth.SMTPPwd != nil &&
		region.SMTPAuth.SMTPPwd.Content != nil && region.SMTPAuth.SMTPPwd.IV == nil {
		encrypted, err := utils.Encrypt(*region.SMTPAuth.SMTPPwd.Content)
		if err != nil {
			return fmt.Errorf("failed to encrypt SMTP password: %v", err)
		}
		region.SMTPAuth.SMTPPwd = encrypted
	}

	return nil
}

// generateRandomToken generates a random token for use in various security contexts
func (s *RegionService) generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// clearRegionCache clears all region-related cache keys
func (s *RegionService) clearRegionCache() {
	if s.cache != nil {
		s.cache.Delete("regions")
		s.cache.Delete("tenant-assignable-regions-info")
	}
}