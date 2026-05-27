package regions

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// RegionService manages geographic region CRUD and association operations.
type RegionService struct {
	regionRepo             repositories.Repository[models.Region]
	tenantConfigDetailRepo repositories.Repository[models.TenantConfigDetail]
	cache                  *database.RedisCache
	db                     *gorm.DB
}

// NewRegionService constructs a RegionService with injected dependencies.
func NewRegionService(db *gorm.DB, cache *database.RedisCache) *RegionService {
	return &RegionService{
		regionRepo:             repositories.NewRepository[models.Region](db),
		tenantConfigDetailRepo: repositories.NewRepository[models.TenantConfigDetail](db),
		cache:                  cache,
		db:                     db,
	}
}

// NewRegionServiceFromGlobal is a convenience wrapper for legacy callers that
// have not yet switched to constructor injection.
func NewRegionServiceFromGlobal() *RegionService {
	return NewRegionService(database.DB, database.Cache)
}

/* ── CREATE ─────────────────────────────────────────────── */

func (s *RegionService) Create(ctx context.Context, createRegionDto *dto.CreateRegionDto) (*models.Region, error) {
	region := &models.Region{}
	if err := copier.Copy(region, createRegionDto); err != nil {
		return nil, fmt.Errorf("failed to map DTO: %w", err)
	}
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	region, err := s.regionRepo.Create(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}
	s.clearCache()
	return region, nil
}

func (s *RegionService) InsertRegions(ctx context.Context, regions *[]dto.CreateRegionDto) (*[]models.Region, error) {
	var models_ []models.Region
	if err := copier.Copy(&models_, regions); err != nil {
		return nil, err
	}
	for i := range models_ {
		if err := s.encryptSensitiveData(&models_[i]); err != nil {
			return nil, fmt.Errorf("failed to encrypt region data: %w", err)
		}
	}
	if err := s.db.Create(&models_).Error; err != nil {
		return nil, fmt.Errorf("failed to bulk insert regions: %w", err)
	}
	s.clearCache()
	return &models_, nil
}

/* ── UPDATE ─────────────────────────────────────────────── */

func (s *RegionService) Update(ctx context.Context, regionId uint, updateRegionDto *dto.UpdateRegionDto) (*models.Region, error) {
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}
	if err := copier.Copy(region, updateRegionDto); err != nil {
		return nil, fmt.Errorf("failed to map DTO: %w", err)
	}
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	if err := s.regionRepo.Update(region); err != nil {
		return nil, err
	}
	s.clearCache()
	return region, nil
}

func (s *RegionService) Save(ctx context.Context, region *models.Region) (*models.Region, error) {
	if err := s.encryptSensitiveData(region); err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	region, err := s.regionRepo.Save(region)
	if err != nil {
		return nil, fmt.Errorf("failed to save region: %w", err)
	}
	s.clearCache()
	return region, nil
}

/* ── READ ────────────────────────────────────────────────── */

func (s *RegionService) FindAllWithOptions(ctx context.Context, findOptions map[string]any) ([]models.Region, int64, error) {
	regions, total, err := s.regionRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get regions: %w", err)
	}
	return regions, total, nil
}

func (s *RegionService) GetAllRegions(ctx context.Context) ([]models.Region, error) {
	if s.cache != nil {
		var cached []models.Region
		if ok, _ := s.cache.Get("regions", &cached); ok {
			return cached, nil
		}
	}
	regions, err := s.regionRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}
	if s.cache != nil {
		s.cache.Set("regions", regions, 25000)
	}
	return regions, nil
}

func (s *RegionService) FindOne(ctx context.Context, id uint) (*models.Region, error) {
	region, err := s.regionRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}
	return region, nil
}

// FindByRegionName satisfies the tenants.RegionLookup interface.
func (s *RegionService) FindByRegionName(ctx context.Context, regionName string) (*models.Region, error) {
	var region models.Region
	err := s.regionRepo.CreateQueryBuilder().
		Where("name = ?", regionName).
		First(&region).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("region not found: %s", regionName)
		}
		return nil, fmt.Errorf("failed to find region by name: %w", err)
	}
	return &region, nil
}

func (s *RegionService) GetTenantAssignableRegionsInfo(ctx context.Context) ([]models.Region, error) {
	if s.cache != nil {
		var cached []models.Region
		if ok, _ := s.cache.Get("tenant-assignable-regions-info", &cached); ok {
			return cached, nil
		}
	}
	var regions []models.Region
	if err := s.regionRepo.CreateQueryBuilder().
		Select("id, name, root_domain_name, description, country, city, tenant_count_capacity").
		Find(&regions).Error; err != nil {
		return nil, fmt.Errorf("failed to get assignable regions: %w", err)
	}
	if s.cache != nil {
		s.cache.Set("tenant-assignable-regions-info", regions, 25000)
	}
	return regions, nil
}

/* ── DELETE ─────────────────────────────────────────────── */

func (s *RegionService) Delete(ctx context.Context, regionId uint) error {
	if err := s.regionRepo.Delete(regionId); err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}
	return nil
}

/* ── ASSOCIATIONS ────────────────────────────────────────── */

func (s *RegionService) AddTenantConfigDetailById(ctx context.Context, regionId, tenantConfigDetailId uint) error {
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}
	tcd, err := s.tenantConfigDetailRepo.FindByID(tenantConfigDetailId)
	if err != nil {
		return fmt.Errorf("failed to find tenant config detail: %w", err)
	}
	tcd.RegionID = regionId
	tcd.Region = *region
	if err := s.tenantConfigDetailRepo.Update(tcd); err != nil {
		return fmt.Errorf("failed to associate tenant config detail: %w", err)
	}
	s.clearCache()
	return nil
}

func (s *RegionService) AddTenantConfigDetailsById(ctx context.Context, regionId uint, tenantConfigDetailIds []uint) ([]models.TenantConfigDetail, error) {
	region, err := s.regionRepo.FindByID(regionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}

	var tcds []models.TenantConfigDetail
	if len(tenantConfigDetailIds) > 0 {
		if err := s.tenantConfigDetailRepo.CreateQueryBuilder().
			Where("id IN ?", tenantConfigDetailIds).Find(&tcds).Error; err != nil {
			return nil, fmt.Errorf("failed to find tenant config details: %w", err)
		}
		if len(tcds) == 0 {
			return nil, fmt.Errorf("no tenant config details found for provided IDs")
		}
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := range tcds {
		tcds[i].RegionID = regionId
		tcds[i].Region = *region
		if err := tx.Save(&tcds[i]).Error; err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				return nil, fmt.Errorf("tenant config detail already assigned to region: %w", err)
			}
			return nil, fmt.Errorf("failed to associate tenant config detail: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var result []models.TenantConfigDetail
	if err := s.tenantConfigDetailRepo.CreateQueryBuilder().
		Where("region_id = ?", regionId).Find(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated config details: %w", err)
	}
	s.clearCache()
	return result, nil
}

func (s *RegionService) RemoveTenantConfigDetailById(ctx context.Context, regionId, tenantConfigDetailId uint) error {
	tcd, err := s.tenantConfigDetailRepo.FindByID(tenantConfigDetailId)
	if err != nil {
		return fmt.Errorf("failed to find tenant config detail: %w", err)
	}
	if tcd.RegionID != regionId {
		return fmt.Errorf("tenant config detail is not assigned to this region")
	}
	tcd.RegionID = 0
	tcd.Region = models.Region{}
	if err := s.tenantConfigDetailRepo.Update(tcd); err != nil {
		return fmt.Errorf("failed to remove association: %w", err)
	}
	s.clearCache()
	return nil
}

func (s *RegionService) RemoveTenantConfigDetailsById(ctx context.Context, regionId uint, tenantConfigDetailIds []uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var tcds []models.TenantConfigDetail
	if err := tx.Where("id IN ? AND region_id = ?", tenantConfigDetailIds, regionId).Find(&tcds).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find tenant config details: %w", err)
	}
	if len(tcds) == 0 {
		tx.Rollback()
		return nil
	}
	for i := range tcds {
		tcds[i].RegionID = 0
		tcds[i].Region = models.Region{}
		if err := tx.Save(&tcds[i]).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove association for %d: %w", tcds[i].ID, err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	s.clearCache()
	return nil
}

/* ── PRIVATE HELPERS ─────────────────────────────────────── */

func (s *RegionService) encryptSensitiveData(region *models.Region) error {
	// Helper: only encrypts if content exists and IV is not yet set (not already encrypted).
	encryptField := func(pwd *struct {
		IV      *string `json:"iv"`
		Content *string `json:"content"`
	}) (*struct {
		IV      *string `json:"iv"`
		Content *string `json:"content"`
	}, error) {
		if pwd == nil || pwd.Content == nil || pwd.IV != nil {
			return pwd, nil
		}
		return utils.Encrypt(*pwd.Content)
	}

	if region.WebServerProperties != nil && region.WebServerProperties.Password != nil {
		enc, err := encryptField(region.WebServerProperties.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt webserver password: %w", err)
		}
		region.WebServerProperties.Password = enc
	}
	if region.DBProperties.Password != nil {
		enc, err := encryptField(region.DBProperties.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt DB password: %w", err)
		}
		region.DBProperties.Password = enc
	}
	if region.ElasticSearchProperties != nil && region.ElasticSearchProperties.Password != nil {
		enc, err := encryptField(region.ElasticSearchProperties.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt elasticsearch password: %w", err)
		}
		region.ElasticSearchProperties.Password = enc
	}
	if region.RedisProperties != nil && region.RedisProperties.Password != nil {
		enc, err := encryptField(region.RedisProperties.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt redis password: %w", err)
		}
		region.RedisProperties.Password = enc
	}
	if region.RootFileSystem != nil && region.RootFileSystem.Password != nil {
		enc, err := encryptField(region.RootFileSystem.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt filesystem password: %w", err)
		}
		region.RootFileSystem.Password = enc
	}
	if region.SMTPAuth != nil && region.SMTPAuth.SMTPPwd != nil {
		enc, err := encryptField(region.SMTPAuth.SMTPPwd)
		if err != nil {
			return fmt.Errorf("failed to encrypt SMTP password: %w", err)
		}
		region.SMTPAuth.SMTPPwd = enc
	}
	return nil
}

func (s *RegionService) clearCache() {
	if s.cache == nil {
		return
	}
	if err := s.cache.Delete("regions"); err != nil {
		slog.Warn("failed to clear regions cache", "error", err)
	}
	if err := s.cache.Delete("tenant-assignable-regions-info"); err != nil {
		slog.Warn("failed to clear assignable-regions cache", "error", err)
	}
}
