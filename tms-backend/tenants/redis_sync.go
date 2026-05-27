package tenants

import (
	"context"
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
	"github.com/redis/go-redis/v9"
)

// SetTenantPropertiesInRedisByRegionName is called on startup or manual recovery
// to re-synchronise all active tenants in a region into the SaaS Redis.
// NOTE: this can be a long-running operation for regions with many tenants.
func (s *TenantService) SetTenantPropertiesInRedisByRegionName(ctx context.Context, regionName string) error {
	region, err := s.regionService.FindByRegionName(ctx, regionName)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}

	tenants, err := s.FindActiveTenantsByRegionName(ctx, regionName)
	if err != nil {
		return fmt.Errorf("failed to find active tenants: %w", err)
	}

	for _, tenant := range tenants {
		// Load config detail so resolveTenantProperties can access it.
		var fullTenant models.Tenant
		if err := s.tenantRepo.CreateQueryBuilder().
			Preload("TenantConfigDetail").
			Preload("PrimaryContact").
			Where("id = ?", tenant.ID).
			First(&fullTenant).Error; err != nil {
			warnErr(fmt.Sprintf("load tenant %d for redis sync", tenant.ID), err)
			continue
		}

		props, err := resolveTenantProperties(ctx, fullTenant, *region)
		if err != nil {
			warnErr(fmt.Sprintf("resolve properties for tenant %d", tenant.ID), err)
			continue
		}

		redisClient, err := s.openRedisForProps(ctx, props)
		if err != nil {
			warnErr(fmt.Sprintf("open redis for tenant %d", tenant.ID), err)
			continue
		}

		writeTenantRedisKeys(ctx, redisClient, props, props.TenantUniquePrefix, "", "", fullTenant)
	}
	return nil
}

// SetTenantPropertiesInRedisByTenantID updates the SaaS Redis keys for a single tenant.
// Call this after a tenant's configuration is created or updated.
func (s *TenantService) SetTenantPropertiesInRedisByTenantID(ctx context.Context, tenantID uint) error {
	var tenant models.Tenant
	if err := s.tenantRepo.CreateQueryBuilder().
		Preload("TenantConfigDetail").
		Preload("PrimaryContact").
		Where("id = ?", tenantID).
		First(&tenant).Error; err != nil {
		return fmt.Errorf("failed to load tenant: %w", err)
	}

	region, err := s.regionService.FindByRegionName(ctx, tenant.RegionName)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}

	props, err := resolveTenantProperties(ctx, tenant, *region)
	if err != nil {
		return fmt.Errorf("failed to resolve tenant properties: %w", err)
	}

	redisClient, err := s.openRedisForProps(ctx, props)
	if err != nil {
		return fmt.Errorf("failed to open Redis connection: %w", err)
	}

	writeTenantRedisKeys(ctx, redisClient, props, props.TenantUniquePrefix, "", "", tenant)
	return nil
}

// UnsetTenantPropertiesInRedisByTenantID removes all SaaS Redis keys for a tenant.
// Call this when deleting or hard-disabling a tenant.
func (s *TenantService) UnsetTenantPropertiesInRedisByTenantID(ctx context.Context, tenantID uint) error {
	var tenant models.Tenant
	if err := s.tenantRepo.CreateQueryBuilder().
		Preload("TenantConfigDetail").
		Where("id = ?", tenantID).
		First(&tenant).Error; err != nil {
		return fmt.Errorf("failed to load tenant: %w", err)
	}

	region, err := s.regionService.FindByRegionName(ctx, tenant.RegionName)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}

	props, err := resolveTenantProperties(ctx, tenant, *region)
	if err != nil {
		return fmt.Errorf("failed to resolve tenant properties: %w", err)
	}

	redisClient, err := s.openRedisForProps(ctx, props)
	if err != nil {
		return fmt.Errorf("failed to open Redis connection: %w", err)
	}

	prefix := props.TenantUniquePrefix
	keys := buildTenantRedisKeyList(props.UniqueName, props.CustomURLSlug, prefix)
	if len(keys) > 0 {
		if err := redisClient.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete tenant Redis keys: %w", err)
		}
	}
	return nil
}

// SetTenantStatusInRedisByTenantID updates only the STATUS key for a tenant in Redis.
// Call this when a tenant is suspended, reactivated, or marked as owing.
func (s *TenantService) SetTenantStatusInRedisByTenantID(ctx context.Context, tenantID uint, status global.TenantStatus) error {
	var tenant models.Tenant
	if err := s.tenantRepo.CreateQueryBuilder().
		Preload("TenantConfigDetail").
		Where("id = ?", tenantID).
		First(&tenant).Error; err != nil {
		return fmt.Errorf("failed to load tenant: %w", err)
	}

	region, err := s.regionService.FindByRegionName(ctx, tenant.RegionName)
	if err != nil {
		return fmt.Errorf("failed to find region: %w", err)
	}

	props, err := resolveTenantProperties(ctx, tenant, *region)
	if err != nil {
		return fmt.Errorf("failed to resolve tenant properties: %w", err)
	}

	redisClient, err := s.openRedisForProps(ctx, props)
	if err != nil {
		return fmt.Errorf("failed to open Redis connection: %w", err)
	}

	return redisClient.Set(ctx, props.TenantUniquePrefix+"STATUS", string(status), 0).Err()
}

// openRedisForProps decrypts the Redis password from TenantProperties and
// returns a client from the pool.
func (s *TenantService) openRedisForProps(ctx context.Context, props *TenantProperties) (*redis.Client, error) {
	if props.RedisProperties == nil {
		return nil, fmt.Errorf("no Redis properties configured for tenant %s", props.TenantUniquePrefix)
	}

	var password string
	if props.RedisProperties.Password != nil {
		decrypted, err := utils.Decrypt(&struct {
			IV      string `json:"iv"`
			Content string `json:"content"`
		}{
			IV:      strOrEmpty(props.RedisProperties.Password.IV),
			Content: strOrEmpty(props.RedisProperties.Password.Content),
		})
		if err != nil {
			warnErr("decrypt redis password for "+props.TenantUniquePrefix, err)
		} else {
			password = decrypted
		}
	}

	cfg := RedisConnectionConfig{
		Host:     props.RedisProperties.Host,
		Port:     props.RedisProperties.Port,
		Password: password,
	}
	return s.redisPool.Get(props.RedisClientName, cfg)
}

// buildTenantRedisKeyList returns the full list of Redis keys that belong to a tenant
// so they can be deleted in one DEL command.
func buildTenantRedisKeyList(uniqueName, customURLSlug, prefix string) []string {
	keys := []string{uniqueName}
	if customURLSlug != "" {
		keys = append(keys, customURLSlug)
	}
	suffixes := []string{
		"NEW_TENANT", "SUPERADMIN_PASSWORD", "SUPERADMIN_PASSWORD_CRYPTO_IV",
		"SUPERADMIN_EMAIL", "SUPERADMIN_FIRSTNAME", "SUPERADMIN_LASTNAME",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD",
		"POSTGRES_PASSWORD_CRYPTO_IV", "POSTGRES_DB", "POSTGRES_SCHEMA",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_PASSWORD_CRYPTO_IV",
		"SMTP_USER", "SMTP_HOST", "SMTP_PORT", "SMTP_SERVICE", "SMTP_SECURE",
		"SMTP_OAUTH", "SMTP_PWORD", "SMTP_PWORD_CRYPTO_IV",
		"JwtConstants_SECRET", "JwtConstants_REFRESH_SECRET",
		"AuthEnabled_Google", "AuthEnabled_Facebook", "AuthEnabled_TwoFactor",
		"FBConstants_APP_ID", "FBConstants_APP_SECRET",
		"GoogleConstants_GOOGLE_OAUTH2_CLIENT_ID", "GoogleConstants_GOOGLE_OAUTH2_CLIENT_SECRET",
		"ELASTICSEARCH_NODE", "ELASTICSEARCH_USERNAME", "ELASTICSEARCH_PASSWORD",
		"LOGO_FILE_SIZE_LIMIT", "PHOTO_FILE_SIZE_LIMIT", "GENERAL_FILE_SIZE_LIMIT",
		"Theme_Custom", "Theme_Type", "Theme_RootUrl",
		"Root_FileSystem_Path", "Root_FileSystem_Username",
		"Root_FileSystem_Password", "Root_FileSystem_Password_CRYPTO_IV",
		"Logo_FileName", "Logo_Mimetype",
		"NAME", "MORE_INFO", "ADDRESS", "STATUS",
		"ResetPasswordMailOptionSettings_TextTemplate",
		"ConfirmEmailMailOptionSettings_TextTemplate",
		"PASSWORD_RESET_EXPIRATION", "EMAIL_VERIFICATION_EXPIRATION",
	}
	for _, s := range suffixes {
		keys = append(keys, prefix+s)
	}
	return keys
}
