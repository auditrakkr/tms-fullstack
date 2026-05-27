package tenants

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dtos "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TenantService manages multi-tenant lifecycle.
type TenantService struct {
	tenantRepo               repositories.Repository[models.Tenant]
	tenantTeamRepo           repositories.Repository[models.TenantTeam]
	tenantAccountOfficerRepo repositories.Repository[models.TenantAccountOfficer]
	customThemeRepo          repositories.Repository[models.CustomTheme]
	themeRepo                repositories.Repository[models.Theme]
	billingRepo              repositories.Repository[models.Billing]
	tenantConfigDetailRepo   repositories.Repository[models.TenantConfigDetail]
	userRepo                 repositories.Repository[models.User]
	regionService            RegionLookup // interface defined in ports.go
	db                       *gorm.DB
	dbPool                   *dbPool          // per-tenant DB connection pool (Phase 2)
	redisPool                *redisClientPool // per-tenant Redis client pool
}

// NewTenantService wires dependencies via DI.
func NewTenantService(db *gorm.DB, regionService RegionLookup) *TenantService {
	return &TenantService{
		tenantRepo:               repositories.NewRepository[models.Tenant](db),
		tenantTeamRepo:           repositories.NewRepository[models.TenantTeam](db),
		tenantAccountOfficerRepo: repositories.NewRepository[models.TenantAccountOfficer](db),
		customThemeRepo:          repositories.NewRepository[models.CustomTheme](db),
		themeRepo:                repositories.NewRepository[models.Theme](db),
		billingRepo:              repositories.NewRepository[models.Billing](db),
		tenantConfigDetailRepo:   repositories.NewRepository[models.TenantConfigDetail](db),
		userRepo:                 repositories.NewRepository[models.User](db),
		regionService:            regionService,
		db:                       db,
		dbPool:                   newDBPool(),
		redisPool:                newRedisClientPool(),
	}
}

// NewTenantServiceFromGlobal is a convenience wrapper for callers that have
// not yet been updated to constructor injection.
func NewTenantServiceFromGlobal(regionService RegionLookup) *TenantService {
	return NewTenantService(database.DB, regionService)
}

/* ── CREATE ─────────────────────────────────────────────── */

func (s *TenantService) CreateTenant(ctx context.Context, createTenantDto *dtos.CreateTenantDto, createPrimaryContact uint) (*models.Tenant, error) {
	newTenant := mapTenantFromDTO(createTenantDto)

	region, err := s.regionService.FindByRegionName(ctx, createTenantDto.RegionName)
	if err != nil {
		return nil, fmt.Errorf("failed to find region: %w", err)
	}
	newTenant.RegionRootDomain = region.RootDomainName

	if createPrimaryContact == 1 {
		// Create a new primary contact user and set as primary contact.
		if createTenantDto.PrimaryContact == nil {
			return nil, global.BadRequest("primaryContact is required when createPrimaryContact=1")
		}
		newUser := &models.User{}
		if err := copier.Copy(newUser, createTenantDto.PrimaryContact); err != nil {
			return nil, fmt.Errorf("failed to map primary contact: %w", err)
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		newUser.PasswordHash = string(hashed)
		newUser.IsPasswordChangeRequired = true
		user, err := s.userRepo.Create(newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create primary contact: %w", err)
		}
		newTenant.PrimaryContactID = user.ID
	} else {
		if createTenantDto.PrimaryContact == nil {
			return nil, global.BadRequest("primaryContact email is required when createPrimaryContact=0")
		}
		primaryContact, err := s.userRepo.FindOne(map[string]any{
			"primary_email_address": createTenantDto.PrimaryContact.PrimaryEmailAddress,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find primary contact: %w", err)
		}
		newTenant.PrimaryContactID = primaryContact.ID
	}

	tenant, err := s.tenantRepo.Save(newTenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	invalidateTenantCache()
	return tenant, nil
}

func (s *TenantService) InsertTenants(ctx context.Context, createDtos []dtos.CreateTenantDto) ([]models.Tenant, error) {
	var tenants []models.Tenant
	for i := range createDtos {
		region, err := s.regionService.FindByRegionName(ctx, createDtos[i].RegionName)
		if err != nil {
			return nil, fmt.Errorf("region not found for tenant %d: %w", i, err)
		}
		t := mapTenantFromDTO(&createDtos[i])
		t.RegionRootDomain = region.RootDomainName
		tenants = append(tenants, *t)
	}
	if err := s.db.Create(&tenants).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, global.Conflict("one or more tenants conflict with existing records")
		}
		return nil, fmt.Errorf("failed to bulk insert tenants: %w", err)
	}
	return tenants, nil
}

/* ── UPDATE ─────────────────────────────────────────────── */

func (s *TenantService) Update(ctx context.Context, tenantId uint, updateTenantDto *dtos.UpdateTenantDto) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}
	if updateTenantDto.Name != nil {
		tenant.Name = *updateTenantDto.Name
	}
	if updateTenantDto.Address != nil {
		tenant.Address = *updateTenantDto.Address
	}
	if updateTenantDto.MoreInfo != nil {
		tenant.MoreInfo = *updateTenantDto.MoreInfo
	}
	if updateTenantDto.Logo != nil {
		tenant.Logo = *updateTenantDto.Logo
	}
	if updateTenantDto.LogoMimeType != nil {
		tenant.LogoMimeType = *updateTenantDto.LogoMimeType
	}
	if updateTenantDto.Status != nil {
		tenant.Status = *updateTenantDto.Status
	}
	if err := s.tenantRepo.Update(tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	invalidateTenantCache()
	return tenant, nil
}

/* ── DELETE ─────────────────────────────────────────────── */

func (s *TenantService) Delete(ctx context.Context, tenantId uint) error {
	if err := s.tenantRepo.Delete(tenantId); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	invalidateTenantCache()
	return nil
}

func (s *TenantService) Remove(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	removed, err := s.tenantRepo.Remove(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to remove tenant: %w", err)
	}
	return &removed, nil
}

/* ── READ ────────────────────────────────────────────────── */

func (s *TenantService) GetAllTenants(ctx context.Context) ([]models.Tenant, error) {
	// Check Redis cache first (reuses the global cache if available).
	if database.Cache != nil {
		var cached []models.Tenant
		if ok, _ := database.Cache.Get("tenants", &cached); ok {
			return cached, nil
		}
	}
	tenants, err := s.tenantRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tenants: %w", err)
	}
	if database.Cache != nil {
		database.Cache.Set("tenants", tenants, 25000)
	}
	return tenants, nil
}

func invalidateTenantCache() {
	if database.Cache != nil {
		database.Cache.Delete("tenants") //nolint:errcheck
	}
}

func (s *TenantService) FindAllWithOptions(ctx context.Context, findOptions map[string]any) ([]models.Tenant, int64, error) {
	tenants, total, err := s.tenantRepo.FindAndCount(findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tenants with options: %w", err)
	}
	return tenants, total, nil
}

func (s *TenantService) FindOne(ctx context.Context, tenantId uint) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}
	return tenant, nil
}

func (s *TenantService) FindActiveTenantsByRegionName(ctx context.Context, regionName string) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().
		Where("region_name = ? AND status = ?", regionName, "active").
		Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active tenants: %w", err)
	}
	return tenants, nil
}

func (s *TenantService) FindTenantsByRegionName(ctx context.Context, regionName string) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().
		Where("region_name = ?", regionName).
		Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find tenants by region: %w", err)
	}
	return tenants, nil
}

func (s *TenantService) FindByActive(ctx context.Context, active bool) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().
		Where("active = ?", active).
		Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find tenants by active status: %w", err)
	}
	return tenants, nil
}

func (s *TenantService) FindByUniqueName(ctx context.Context, subDomain, rootDomain string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().
		Where("subdomain = ? AND region_root_domain = ?", subDomain, rootDomain).
		First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("tenant not found")
		}
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}
	return &tenant, nil
}

func (s *TenantService) CheckDomainExists(ctx context.Context, subDomain, rootDomain string) (bool, error) {
	var count int64
	err := s.tenantRepo.CreateQueryBuilder().
		Where("subdomain = ? AND region_root_domain = ?", subDomain, rootDomain).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check domain: %w", err)
	}
	return count > 0, nil
}

func (s *TenantService) GetActiveTenantsByAccountOfficer(ctx context.Context, userID uint) ([]models.TenantAccountOfficer, error) {
	var officers []models.TenantAccountOfficer
	err := s.tenantAccountOfficerRepo.CreateQueryBuilder().
		Joins("JOIN tenants ON tenants.id = tenant_account_officers.tenant_id").
		Where("tenant_account_officers.user_id = ? AND tenants.active = true AND tenants.deleted_at IS NULL", userID).
		Preload("Tenant").
		Find(&officers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account officer tenants: %w", err)
	}
	return officers, nil
}

func (s *TenantService) GetAccountOfficersByTenant(ctx context.Context, tenantID uint) ([]models.TenantAccountOfficer, error) {
	var officers []models.TenantAccountOfficer
	err := s.tenantAccountOfficerRepo.CreateQueryBuilder().
		Where("tenant_id = ?", tenantID).
		Preload("User").
		Find(&officers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account officers for tenant: %w", err)
	}
	return officers, nil
}

func (s *TenantService) GetTenantIdentities(ctx context.Context, tenantID uint) (*TenantIdentity, error) {
	var tenant models.Tenant
	err := s.tenantRepo.CreateQueryBuilder().
		Preload("PrimaryContact").
		Where("id = ?", tenantID).
		First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, global.NotFound("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant identities: %w", err)
	}
	identity := &TenantIdentity{
		UUID:          tenant.UUID.String(),
		SubDomainName: tenant.Subdomain,
		CustomURLSlug: tenant.CustomURLSlug,
		RegionName:    tenant.RegionName,
		Name:          tenant.Name,
		MoreInfo:      tenant.MoreInfo,
		Address:       tenant.Address,
		Status:        tenant.Status,
	}
	identity.PrimaryContact.FirstName = tenant.PrimaryContact.FirstName
	identity.PrimaryContact.LastName = tenant.PrimaryContact.LastName
	identity.PrimaryContact.PrimaryEmailAddress = tenant.PrimaryContact.PrimaryEmailAddress
	identity.PrimaryContact.Gender = tenant.PrimaryContact.Gender
	identity.PrimaryContact.DateOfBirth = tenant.PrimaryContact.DateOfBirth
	return identity, nil
}

/* ── PRIMARY CONTACT ─────────────────────────────────────── */

func (s *TenantService) CreateAndSetPrimaryContact(ctx context.Context, tenantID uint, createUserDto *dtos.CreateUserDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		newUser := &models.User{}
		if err := copier.Copy(newUser, createUserDto); err != nil {
			return fmt.Errorf("failed to map user dto: %w", err)
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		newUser.PasswordHash = string(hashed)
		newUser.IsPasswordChangeRequired = true
		if err := tx.Create(newUser).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				return global.Conflict("user with this email already exists")
			}
			return fmt.Errorf("failed to create primary contact: %w", err)
		}
		return tx.Model(&models.Tenant{}).Where("id = ?", tenantID).
			Update("primary_contact_id", newUser.ID).Error
	})
}

func (s *TenantService) SetPrimaryContactByID(ctx context.Context, tenantID, userID uint) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", tenantID).
		Update("primary_contact_id", userID).Error
}

func (s *TenantService) UnsetPrimaryContactByID(ctx context.Context, tenantID uint) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", tenantID).
		Update("primary_contact_id", nil).Error
}

/* ── TENANT TEAM ─────────────────────────────────────────── */

func (s *TenantService) CreateAndSetTeamMember(ctx context.Context, tenantID uint, dto *dtos.CreateTenantTeamDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		team := &models.TenantTeam{
			TenantID:         tenantID,
			Roles:            dto.Roles,
			TenantUniqueName: dto.TenantUniqueName,
			TenantUniqueID:   uint(dto.TenantUniqueId),
		}
		if dto.User != nil {
			newUser := &models.User{}
			if err := copier.Copy(newUser, dto.User); err != nil {
				return fmt.Errorf("failed to map user dto: %w", err)
			}
			hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			newUser.PasswordHash = string(hashed)
			if err := tx.Create(newUser).Error; err != nil {
				return fmt.Errorf("failed to create team member user: %w", err)
			}
			team.UserID = newUser.ID
		}
		return tx.Create(team).Error
	})
}

func (s *TenantService) SetTeamMemberByID(ctx context.Context, tenantID, userID uint, roles *TenantTeamRoles) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		team := &models.TenantTeam{
			TenantID: tenantID,
			UserID:   userID,
			Roles:    roles.Roles,
		}
		return tx.Create(team).Error
	})
}

func (s *TenantService) DeleteTeamMemberByID(ctx context.Context, tenantID, userID uint) error {
	return s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Delete(&models.TenantTeam{}).Error
}

func (s *TenantService) UpdateTeamMemberByID(ctx context.Context, tenantID, userID uint, roles *TenantTeamRoles) error {
	return s.db.Model(&models.TenantTeam{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("roles", roles.Roles).Error
}

/* ── ACCOUNT OFFICERS ────────────────────────────────────── */

func (s *TenantService) CreateAndSetTenantAccountOfficer(ctx context.Context, tenantID uint, dto *dtos.CreateTenantAccountOfficerDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		officer := &models.TenantAccountOfficer{
			TenantID: tenantID,
			Roles:    dto.Roles,
		}
		if dto.User != nil {
			newUser := &models.User{}
			if err := copier.Copy(newUser, dto.User); err != nil {
				return fmt.Errorf("failed to map user dto: %w", err)
			}
			hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.PasswordHash), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			newUser.PasswordHash = string(hashed)
			if err := tx.Create(newUser).Error; err != nil {
				return fmt.Errorf("failed to create account officer user: %w", err)
			}
			officer.UserID = newUser.ID
		}
		return tx.Create(officer).Error
	})
}

func (s *TenantService) SetTenantAccountOfficerByID(ctx context.Context, tenantID, userID uint, roles *TenantAccountOfficerRoles) error {
	officer := &models.TenantAccountOfficer{
		TenantID: tenantID,
		UserID:   userID,
		Roles:    roles.Roles,
	}
	return s.db.Create(officer).Error
}

func (s *TenantService) DeleteTenantAccountOfficerByID(ctx context.Context, tenantID, userID uint) error {
	return s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Delete(&models.TenantAccountOfficer{}).Error
}

func (s *TenantService) UpdateTenantAccountOfficerByID(ctx context.Context, tenantID, userID uint, roles *TenantAccountOfficerRoles) error {
	return s.db.Model(&models.TenantAccountOfficer{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("roles", roles.Roles).Error
}

/* ── CUSTOM THEME ────────────────────────────────────────── */

func (s *TenantService) CreateAndSetCustomTheme(ctx context.Context, tenantID uint, dto *dtos.CreateCustomThemeDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		theme := &models.CustomTheme{
			Name:        dto.Name,
			TenantID:    tenantID,
		}
		if dto.Description != nil {
			theme.Description = *dto.Description
		}
		if dto.Properties != nil {
			theme.Properties = *dto.Properties
		}
		return tx.Create(theme).Error
	})
}

func (s *TenantService) SetCustomThemeByID(ctx context.Context, tenantID, customThemeID uint) error {
	return s.db.Model(&models.CustomTheme{}).
		Where("id = ?", customThemeID).
		Update("tenant_id", tenantID).Error
}

func (s *TenantService) UnsetCustomThemeByID(ctx context.Context, tenantID uint) error {
	return s.db.Model(&models.CustomTheme{}).
		Where("tenant_id = ?", tenantID).
		Update("tenant_id", nil).Error
}

/* ── THEME ASSOCIATIONS ──────────────────────────────────── */

func (s *TenantService) AddThemeByID(ctx context.Context, tenantID, themeID uint) error {
	tenant := &models.Tenant{}
	tenant.ID = tenantID
	theme := &models.Theme{}
	theme.ID = themeID
	return s.db.Model(tenant).Association("Themes").Append(theme)
}

func (s *TenantService) RemoveThemeByID(ctx context.Context, tenantID, themeID uint) error {
	tenant := &models.Tenant{}
	tenant.ID = tenantID
	theme := &models.Theme{}
	theme.ID = themeID
	return s.db.Model(tenant).Association("Themes").Delete(theme)
}

/* ── BILLING ASSOCIATIONS ────────────────────────────────── */

func (s *TenantService) CreateAndAddBilling(ctx context.Context, tenantID uint, dto *dtos.CreateBillingDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		billing := &models.Billing{TenantID: tenantID}
		if err := copier.Copy(billing, dto); err != nil {
			return fmt.Errorf("failed to map billing dto: %w", err)
		}
		billing.TenantID = tenantID
		return tx.Create(billing).Error
	})
}

func (s *TenantService) AddBillingByID(ctx context.Context, tenantID, billingID uint) error {
	return s.db.Model(&models.Billing{}).
		Where("id = ?", billingID).
		Update("tenant_id", tenantID).Error
}

func (s *TenantService) RemoveBillingByID(ctx context.Context, tenantID, billingID uint) error {
	return s.db.Model(&models.Billing{}).
		Where("id = ? AND tenant_id = ?", billingID, tenantID).
		Update("tenant_id", nil).Error
}

/* ── LOGO ────────────────────────────────────────────────── */

func (s *TenantService) SetTenantLogo(ctx context.Context, tenantID uint, file *multipart.FileHeader) error {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return fmt.Errorf("failed to find tenant: %w", err)
	}
	if file.Size > global.LOGO_FILE_SIZE_LIMIT {
		return global.BadRequest(fmt.Sprintf("logo too large (max %d bytes)", global.LOGO_FILE_SIZE_LIMIT))
	}

	fileName := file.Filename
	if tenant.Logo != "" {
		fileName = tenant.Logo
	} else {
		fileName = fmt.Sprintf("%d_%s", tenant.ID, file.Filename)
	}

	logoDir := filepath.Join(global.UPLOAD_DIRECTORY, "logos")
	if err := os.MkdirAll(logoDir, 0755); err != nil {
		return fmt.Errorf("failed to create logo directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dest := filepath.Join(logoDir, fileName)
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	tenant.Logo = fileName
	tenant.LogoMimeType = file.Header.Get("Content-Type")
	return s.tenantRepo.Update(tenant)
}

func (s *TenantService) GetTenantLogo(ctx context.Context, tenantID uint) (filePath, mimeType string, err error) {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return "", "", fmt.Errorf("failed to find tenant: %w", err)
	}
	defaultPath := filepath.Join(global.UPLOAD_DIRECTORY, "logos", "blankLogoAvatar.png")
	if tenant.Logo == "" {
		return defaultPath, "image/png", nil
	}
	logoPath := filepath.Join(global.UPLOAD_DIRECTORY, "logos", tenant.Logo)
	if _, statErr := os.Stat(logoPath); os.IsNotExist(statErr) {
		return defaultPath, "image/png", nil
	}
	return logoPath, tenant.LogoMimeType, nil
}

/* ── TENANT CONFIG DETAIL ────────────────────────────────── */

// CreateAndSetTenantConfigDetail is the core tenant provisioner:
//  1. Saves the TenantConfigDetail record and derives the tenantUniquePrefix from UUID.
//  2. Opens (or reuses) a GORM connection to the tenant's database.
//  3. Creates the tenant's PostgreSQL schema.
//  4. Creates the required upload directories under the region's root filesystem.
//  5. Copies avatar placeholder images.
//  6. Opens (or reuses) a Redis client for the region and writes ~50 keys that
//     the SaaS gateway/web-server layer reads at runtime.
//
// The method is idempotent for schema and directory creation (uses IF NOT EXISTS /
// MkdirAll). Redis keys are overwritten on every call.
func (s *TenantService) CreateAndSetTenantConfigDetail(ctx context.Context, tenantID uint, createDto *dtos.CreateTenantConfigDetailDto) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Load tenant + primary contact.
		var tenant models.Tenant
		if err := tx.Preload("PrimaryContact").First(&tenant, tenantID).Error; err != nil {
			return fmt.Errorf("failed to load tenant: %w", err)
		}

		// 2. Compute tenant unique prefix from UUID.
		tenantUniquePrefix := "_" + strings.ReplaceAll(tenant.UUID.String(), "-", "") + "_"

		// 3. Resolve region.
		region, err := s.regionService.FindByRegionName(ctx, tenant.RegionName)
		if err != nil {
			return fmt.Errorf("failed to find region: %w", err)
		}

		// 4. Create and save TenantConfigDetail.
		newDetail := &models.TenantConfigDetail{}
		if err := copier.Copy(newDetail, createDto); err != nil {
			return fmt.Errorf("failed to map config detail dto: %w", err)
		}
		newDetail.DBSchema = tenantUniquePrefix
		newDetail.TenantID = tenantID
		newDetail.RegionID = region.ID
		newDetail.Region = *region

		if err := tx.Create(newDetail).Error; err != nil {
			return fmt.Errorf("failed to save tenant config detail: %w", err)
		}

		// Link to tenant.
		if err := tx.Model(&models.Tenant{}).Where("id = ?", tenantID).
			Update("tenant_config_detail_id", newDetail.ID).Error; err != nil {
			// Not fatal — the TenantID FK on TenantConfigDetail already links them.
			warnErr("link config detail to tenant", err)
		}

		// 5. Resolve effective properties.
		tenant.TenantConfigDetail = *newDetail
		props, err := resolveTenantProperties(ctx, tenant, *region)
		if err != nil {
			return fmt.Errorf("failed to resolve tenant properties: %w", err)
		}

		// 6. Open tenant DB connection and create schema.
		if props.DBProperties != nil && props.DBProperties.Password != nil {
			dbPassword, decErr := utils.Decrypt(&struct {
				IV      string `json:"iv"`
				Content string `json:"content"`
			}{
				IV:      *props.DBProperties.Password.IV,
				Content: *props.DBProperties.Password.Content,
			})
			if decErr != nil {
				warnErr("decrypt db password", decErr)
			} else {
				dbCfg := DBConnectionConfig{
					Host:     props.DBProperties.Host,
					Username: props.DBProperties.Username,
					Password: dbPassword,
					Database: props.DBProperties.Database,
					Port:     props.DBProperties.Port,
				}
				tenantDB, connErr := s.dbPool.Get(props.DBConnectionName, dbCfg)
				if connErr != nil {
					warnErr("open tenant DB connection", connErr)
				} else {
					// Create schema — IF NOT EXISTS makes this idempotent.
					if schemaErr := tenantDB.Exec(
						"CREATE SCHEMA IF NOT EXISTS " + tenantUniquePrefix,
					).Error; schemaErr != nil {
						warnErr("create tenant schema", schemaErr)
					}
				}
			}
		}

		// 7. Create upload directories under the region root filesystem.
		rootFS := ""
		if props.RootFileSystem != nil {
			rootFS = props.RootFileSystem.Path
		}
		if rootFS == "" {
			rootFS = global.UPLOAD_DIRECTORY
		}
		tenantUploadDir := filepath.Join(rootFS, tenantUniquePrefix)
		dirs := []string{
			filepath.Join(tenantUploadDir, "photos", "users"),
			filepath.Join(tenantUploadDir, "photos", "products"),
			filepath.Join(tenantUploadDir, "general"),
			filepath.Join(tenantUploadDir, "logos"),
			filepath.Join(tenantUploadDir, "theme", "css"),
			filepath.Join(tenantUploadDir, "theme", "js"),
			filepath.Join(tenantUploadDir, "theme", "img"),
		}
		for _, d := range dirs {
			if mkErr := os.MkdirAll(d, 0755); mkErr != nil {
				warnErr("create directory "+d, mkErr)
			}
		}

		// 8. Copy avatar placeholders.
		copyAvatar("assets/avatars/blankPhotoAvatar.png",
			filepath.Join(tenantUploadDir, "photos", "users", "blankPhotoAvatar.png"))
		copyAvatar("assets/avatars/blankLogoAvatar.png",
			filepath.Join(tenantUploadDir, "logos", "blankLogoAvatar.png"))

		// 9. Generate superadmin credentials.
		superAdminPassword, genErr := utils.GeneratePassword()
		if genErr != nil {
			warnErr("generate superadmin password", genErr)
		}
		var saPasswordEncrypted, saPasswordIV string
		if superAdminPassword != "" {
			encResult, encErr := utils.Encrypt(superAdminPassword)
			if encErr != nil {
				warnErr("encrypt superadmin password", encErr)
			} else if encResult != nil {
				if encResult.Content != nil {
					saPasswordEncrypted = *encResult.Content
				}
				if encResult.IV != nil {
					saPasswordIV = *encResult.IV
				}
			}
		}

		// 10. Write to SaaS Redis.
		if props.RedisProperties != nil && props.RedisProperties.Password != nil {
			redisPassword, decErr := utils.Decrypt(&struct {
				IV      string `json:"iv"`
				Content string `json:"content"`
			}{
				IV:      *props.RedisProperties.Password.IV,
				Content: *props.RedisProperties.Password.Content,
			})
			if decErr != nil {
				warnErr("decrypt redis password", decErr)
			} else {
				redisCfg := RedisConnectionConfig{
					Host:     props.RedisProperties.Host,
					Port:     props.RedisProperties.Port,
					Password: redisPassword,
				}
				redisClient, rErr := s.redisPool.Get(props.RedisClientName, redisCfg)
				if rErr != nil {
					warnErr("open tenant Redis connection", rErr)
				} else {
					writeTenantRedisKeys(ctx, redisClient, props, tenantUniquePrefix, saPasswordEncrypted, saPasswordIV, tenant)
				}
			}
		}

		return nil
	})
}

func (s *TenantService) SetTenantConfigDetailByID(ctx context.Context, tenantID, configDetailID uint) error {
	return s.db.Model(&models.TenantConfigDetail{}).
		Where("id = ?", configDetailID).
		Update("tenant_id", tenantID).Error
}

func (s *TenantService) UnsetTenantConfigDetailByID(ctx context.Context, tenantID uint) error {
	return s.db.Model(&models.TenantConfigDetail{}).
		Where("tenant_id = ?", tenantID).
		Update("tenant_id", nil).Error
}

/* ── LEGACY REDIS (kept for backward compat during Phase 2 migration) ── */

func (s *TenantService) getRedisClient(name string, opts redis.Options) (*redis.Client, error) {
	return s.redisPool.getByOptions(name, opts)
}

/* ── PRIVATE HELPERS ─────────────────────────────────────── */

func mapTenantFromDTO(dto *dtos.CreateTenantDto) *models.Tenant {
	t := &models.Tenant{}
	if dto.Name != "" {
		t.Name = dto.Name
	}
	if dto.Address != "" {
		t.Address = dto.Address
	}
	if dto.MoreInfo != nil {
		t.MoreInfo = *dto.MoreInfo
	}
	if dto.Logo != nil {
		t.Logo = *dto.Logo
	}
	if dto.LogoMimeType != nil {
		t.LogoMimeType = *dto.LogoMimeType
	}
	if dto.Status != nil {
		t.Status = *dto.Status
	}
	if dto.RegionName != "" {
		t.RegionName = dto.RegionName
	}
	return t
}

// warnErr logs a non-fatal error during cleanup operations.
func warnErr(op string, err error) {
	if err != nil {
		slog.Warn("non-fatal error", "op", op, "error", err)
	}
}

// copyAvatar copies src to dst; logs a warning on failure (non-fatal).
func copyAvatar(src, dst string) {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		slog.Warn("avatar source not found, skipping copy", "src", src)
		return
	}
	in, err := os.Open(src)
	if err != nil {
		warnErr("open avatar "+src, err)
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		warnErr("create avatar "+dst, err)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		warnErr("copy avatar", err)
	}
}

// writeTenantRedisKeys writes the ~50 SaaS configuration keys for a single tenant
// to its region's Redis instance. This allows the web-server / reverse-proxy layer
// to look up tenant configuration without hitting the database.
func writeTenantRedisKeys(
	ctx context.Context,
	client *redis.Client,
	props *TenantProperties,
	prefix, saPasswordEncrypted, saPasswordIV string,
	tenant models.Tenant,
) {
	kv := make(map[string]any)

	// Routing keys
	kv[props.UniqueName] = prefix
	if props.CustomURLSlug != "" {
		kv[props.CustomURLSlug] = prefix
	}

	// New-tenant bootstrap
	kv[prefix+"NEW_TENANT"] = "true"
	kv[prefix+"SUPERADMIN_PASSWORD"] = saPasswordEncrypted
	kv[prefix+"SUPERADMIN_PASSWORD_CRYPTO_IV"] = saPasswordIV
	kv[prefix+"SUPERADMIN_EMAIL"] = props.PrimaryContact.PrimaryEmailAddress
	kv[prefix+"SUPERADMIN_FIRSTNAME"] = props.PrimaryContact.FirstName
	kv[prefix+"SUPERADMIN_LASTNAME"] = props.PrimaryContact.LastName
	kv[prefix+"SUPERADMIN_GENDER"] = string(props.PrimaryContact.Gender)

	// Postgres
	if db := props.DBProperties; db != nil {
		kv[prefix+"POSTGRES_HOST"] = db.Host
		kv[prefix+"POSTGRES_PORT"] = db.Port
		kv[prefix+"POSTGRES_USER"] = db.Username
		kv[prefix+"POSTGRES_DB"] = db.Database
		kv[prefix+"POSTGRES_SCHEMA"] = prefix
		if db.Password != nil {
			kv[prefix+"POSTGRES_PASSWORD"] = strOrEmpty(db.Password.Content)
			kv[prefix+"POSTGRES_PASSWORD_CRYPTO_IV"] = strOrEmpty(db.Password.IV)
		}
	}

	// Redis
	if r := props.RedisProperties; r != nil {
		kv[prefix+"REDIS_HOST"] = r.Host
		kv[prefix+"REDIS_PORT"] = r.Port
		if r.Password != nil {
			kv[prefix+"REDIS_PASSWORD"] = strOrEmpty(r.Password.Content)
			kv[prefix+"REDIS_PASSWORD_CRYPTO_IV"] = strOrEmpty(r.Password.IV)
		}
	}

	// SMTP
	if s := props.SMTPAuth; s != nil {
		kv[prefix+"SMTP_USER"] = s.SMTPUser
		kv[prefix+"SMTP_HOST"] = s.SMTPHost
		kv[prefix+"SMTP_PORT"] = s.SMTPPort
		kv[prefix+"SMTP_SERVICE"] = s.SMTPService
		kv[prefix+"SMTP_SECURE"] = boolStr(s.SMTPSecure == "true")
		kv[prefix+"SMTP_OAUTH"] = boolStr(s.SMTPOauth)
		if s.SMTPPwd != nil {
			kv[prefix+"SMTP_PWORD"] = strOrEmpty(s.SMTPPwd.Content)
			kv[prefix+"SMTP_PWORD_CRYPTO_IV"] = strOrEmpty(s.SMTPPwd.IV)
		}
	}

	// JWT
	if j := props.JWTConstants; j != nil {
		kv[prefix+"JwtConstants_SECRET"] = j.JWTSecretKey
		kv[prefix+"JwtConstants_SECRET_KEY_EXPIRATION"] = j.JWTSecretKeyExpiration
		kv[prefix+"JwtConstants_REFRESH_SECRET"] = j.JWTRefreshSecret
		kv[prefix+"JwtConstants_REFRESH_SECRET_KEY_EXPIRATION"] = j.JWTRefreshSecretKeyExpiration
		kv[prefix+"JwtConstants_SECRET_PRIVATE_KEY"] = j.JWTSecretPrivateKey
		kv[prefix+"JwtConstants_SECRET_PUBLIC_KEY"] = j.JWTSecretPublicKey
		kv[prefix+"JwtConstants_SIGN_ALGORITHM"] = j.JWTSigningAlgorithm
	}

	// Auth enabled
	if a := props.AuthEnabled; a != nil {
		kv[prefix+"AuthEnabled_Google"] = boolStr(a.Google)
		kv[prefix+"AuthEnabled_Facebook"] = boolStr(a.Facebook)
		kv[prefix+"AuthEnabled_TwoFactor"] = boolStr(a.TwoFactor)
	}

	// Facebook
	if fb := props.FBOauth2Constants; fb != nil {
		kv[prefix+"FBConstants_APP_ID"] = fb.FBAppID
		kv[prefix+"FBConstants_APP_SECRET"] = fb.FBAppSecret
		kv[prefix+"FBConstants_CREATE_USER_IF_NOT_EXISTS"] = boolStr(fb.CreateUserIfNotExists)
	}

	// Google
	if g := props.GoogleOidcConstants; g != nil {
		kv[prefix+"GoogleConstants_GOOGLE_OAUTH2_CLIENT_OIDC_ISSUER"] = g.GoogleOauth2ClientOidcIssuer
		kv[prefix+"GoogleConstants_GOOGLE_API_KEY"] = g.GoogleApiKey
		kv[prefix+"GoogleConstants_GOOGLE_OAUTH2_CLIENT_ID"] = g.GoogleOauth2ClientID
		kv[prefix+"GoogleConstants_GOOGLE_OAUTH2_CLIENT_SECRET"] = g.GoogleOauth2ClientSecret
		kv[prefix+"GoogleConstants_CREATE_USER_IF_NOT_EXISTS"] = boolStr(g.CreateUserIfNotExists)
	}

	// Elasticsearch
	if es := props.ElasticSearchProperties; es != nil {
		kv[prefix+"ELASTICSEARCH_NODE"] = es.Node
		kv[prefix+"ELASTICSEARCH_USERNAME"] = es.Username
		if es.Password != nil {
			kv[prefix+"ELASTICSEARCH_PASSWORD"] = strOrEmpty(es.Password.Content)
			kv[prefix+"ELASTICSEARCH_PASSWORD_CRYPTO_IV"] = strOrEmpty(es.Password.IV)
		}
	}

	// Size limits
	if sz := props.SizeLimits; sz != nil {
		kv[prefix+"LOGO_FILE_SIZE_LIMIT"] = sz.LogoFileSizeLimit
		kv[prefix+"PHOTO_FILE_SIZE_LIMIT"] = sz.PhotoFileSizeLimit
		kv[prefix+"GENERAL_FILE_SIZE_LIMIT"] = sz.GeneralFileSizeLimit
	}

	// Theme
	if t := props.Theme; t != nil {
		kv[prefix+"Theme_Custom"] = boolStr(t.Custom)
		kv[prefix+"Theme_Type"] = t.Type
		kv[prefix+"Theme_RootUrl"] = t.RootUrl
	}

	// Root filesystem
	if fs := props.RootFileSystem; fs != nil {
		kv[prefix+"Root_FileSystem_Path"] = fs.Path
		if fs.Username != nil {
			kv[prefix+"Root_FileSystem_Username"] = *fs.Username
		}
		if fs.Password != nil {
			kv[prefix+"Root_FileSystem_Password"] = strOrEmpty(fs.Password.Content)
			kv[prefix+"Root_FileSystem_Password_CRYPTO_IV"] = strOrEmpty(fs.Password.IV)
		}
	}

	// Tenant metadata
	kv[prefix+"NAME"] = props.Name
	kv[prefix+"MORE_INFO"] = props.MoreInfo
	kv[prefix+"ADDRESS"] = props.Address
	kv[prefix+"STATUS"] = string(props.Status)
	kv[prefix+"Logo_FileName"] = props.Logo
	kv[prefix+"Logo_Mimetype"] = props.LogoMimeType

	// Other user options
	if o := props.OtherUserOptions; o != nil {
		kv[prefix+"ResetPasswordMailOptionSettings_TextTemplate"] = o.ResetPasswordMailOptionSettings_TextTemplate
		kv[prefix+"ConfirmEmailMailOptionSettings_TextTemplate"] = o.ConfirmEmailMailOptionSettings_TextTemplate
		kv[prefix+"PASSWORD_RESET_EXPIRATION"] = o.PasswordResetExpiration
		kv[prefix+"EMAIL_VERIFICATION_EXPIRATION"] = o.EmailVerificationExpiration
	}

	// Write all keys in one pipelined MSET.
	pipe := client.Pipeline()
	for k, v := range kv {
		pipe.Set(ctx, k, fmt.Sprintf("%v", v), 0)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		warnErr("Redis MSET for tenant "+prefix, err)
	}
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
