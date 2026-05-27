package tenants

import (
	"context"
	"fmt"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
)

// TenantProperties is the resolved runtime configuration for a single tenant.
// Each field falls back to the region default when the tenant's TenantConfigDetail
// does not override it - mirrors the NestJS getTenantPropertiesByRegionName logic.
type TenantProperties struct {
	// Identity
	TenantUniquePrefix string // e.g. "_a1b2c3d4e5f6_"
	UniqueName         string // e.g. "acme.us-east.example.com"
	CustomURLSlug      string
	Name               string
	MoreInfo           string
	Address            string
	Status             global.TenantStatus
	Logo               string
	LogoMimeType       string

	// Infrastructure -- effective values (tenant overrides region)
	DBProperties            *models.DBProperties
	RedisProperties         *models.RedisProperties
	ElasticSearchProperties *models.ElasticSearchProperties
	RootFileSystem          *models.RootFileSystem
	SMTPAuth                *models.SMTPAuth
	JWTConstants            *models.JWTConstants
	AuthEnabled             *models.AuthEnabled
	FBOauth2Constants       *models.FBOauth2Constants
	GoogleOidcConstants     *models.GoogleOauth2Constants
	OtherUserOptions        *models.OtherUserOptions
	SizeLimits              *models.SizeLimits
	Theme                   *models.ThemeType

	// For Redis client selection
	HasTenantSpecificRedis bool
	RedisClientName        string
	DBConnectionName       string

	// Primary contact - used when seeding the superadmin on new tenant provision
	PrimaryContact struct {
		FirstName           string
		LastName            string
		PrimaryEmailAddress string
		Gender              global.Gender
	}
}

// resolveTenantProperties applies the tenant-overrides-region fallback strategy
// for every infrastructure field and returns a fully-resolved TenantProperties.
// This is called both by createAndSetTenantConfigDetail (single tenant) and by
// SetTenantPropertiesInRedisByRegionName (bulk region sync).
func resolveTenantProperties(ctx context.Context, tenant models.Tenant, region models.Region) (*TenantProperties, error) {
	cd := tenant.TenantConfigDetail

	uuid := tenant.UUID.String()
	prefix := "_" + strings.ReplaceAll(uuid, "-", "") + "_"
	uniqueName := fmt.Sprintf("%s.%s", tenant.Subdomain, tenant.RegionRootDomain)

	p := &TenantProperties{
		TenantUniquePrefix: prefix,
		UniqueName:         uniqueName,
		CustomURLSlug:      tenant.CustomURLSlug,
		Name:               tenant.Name,
		MoreInfo:           tenant.MoreInfo,
		Address:            tenant.Address,
		Status:             tenant.Status,
		Logo:               tenant.Logo,
		LogoMimeType:       tenant.LogoMimeType,
	}

	// Infrastructure - tenant config detail overrides region.
	// DBProperties
	if cd.DBProperties != nil {
		p.DBProperties = cd.DBProperties
	} else {
		p.DBProperties = &region.DBProperties
	}

	// RedisProperties
	if cd.RedisProperties != nil {
		p.RedisProperties = cd.RedisProperties
		p.HasTenantSpecificRedis = true
		p.RedisClientName = tenant.RegionName + "_" + prefix
		p.DBConnectionName = region.Name + "_" + prefix
	} else {
		p.RedisProperties = region.RedisProperties
		p.HasTenantSpecificRedis = false
		p.RedisClientName = tenant.RegionName
		p.DBConnectionName = region.Name
	}

	// ElasticSearch
	if cd.ElasticSearchProperties != nil {
		p.ElasticSearchProperties = cd.ElasticSearchProperties
	} else {
		p.ElasticSearchProperties = region.ElasticSearchProperties
	}

	// RootFileSystem
	if cd.RootFileSystem != nil {
		p.RootFileSystem = cd.RootFileSystem
	} else {
		p.RootFileSystem = region.RootFileSystem
	}

	// SMTP
	if cd.SMTPAuth != nil {
		p.SMTPAuth = cd.SMTPAuth
	} else {
		p.SMTPAuth = region.SMTPAuth
	}

	// JWT
	if cd.JWTConstants != nil {
		p.JWTConstants = cd.JWTConstants
	} else {
		p.JWTConstants = region.JWTConstants
	}

	// Auth enabled flags
	if cd.AuthEnabled != nil {
		p.AuthEnabled = cd.AuthEnabled
	} else {
		p.AuthEnabled = region.AuthEnabled
	}

	// Facebook
	if cd.FBOauth2Constants != nil {
		p.FBOauth2Constants = cd.FBOauth2Constants
	} else {
		p.FBOauth2Constants = region.FBOauth2Constants
	}

	// Google OIDC
	if cd.GoogleOauth2Constants != nil {
		p.GoogleOidcConstants = cd.GoogleOauth2Constants
	} else {
		p.GoogleOidcConstants = region.GoogleOidcConstants
	}

	// Other user options
	if cd.OtherUserOptions != nil {
		p.OtherUserOptions = cd.OtherUserOptions
	} else {
		p.OtherUserOptions = region.OtherUserOptions
	}

	// Size limits
	if cd.SizeLimits != nil {
		p.SizeLimits = cd.SizeLimits
	} else {
		p.SizeLimits = region.SizeLimits
	}

	// Theme
	if cd.Theme != nil {
		p.Theme = cd.Theme
	} else {
		p.Theme = region.Theme
	}

	// Primary contact
	p.PrimaryContact.FirstName = tenant.PrimaryContact.FirstName
	p.PrimaryContact.LastName = tenant.PrimaryContact.LastName
	p.PrimaryContact.PrimaryEmailAddress = tenant.PrimaryContact.PrimaryEmailAddress
	p.PrimaryContact.Gender = tenant.PrimaryContact.Gender

	return p, nil
}
