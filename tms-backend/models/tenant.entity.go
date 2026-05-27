package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	UUID          uuid.UUID          `gorm:"type:uuid;default:uuid_generate_v4();unique;not null"`
	Name          string             `gorm:"type:varchar(255);not null"`
	Subdomain     string             `gorm:"type:varchar(255);uniqueIndex:idx_subdomain_region;not null"`
	CustomURLSlug string             `gorm:"type:varchar(255);unique"`
	Address       string             `gorm:"type:varchar(255)"`
	MoreInfo      string             `gorm:"type:text"`
	Logo          string             `gorm:"type:varchar(255)"`
	LogoMimeType  string             `gorm:"type:varchar(255)"`
	Status        global.TenantStatus `gorm:"type:tenant_status;default:'active'" json:"status"`
	Active        bool               `gorm:"default:false"`

	PrimaryContactID uint
	PrimaryContact   User

	CustomTheme           CustomTheme
	TenantTeams           []TenantTeam
	TenantAccountOfficers []TenantAccountOfficer

	UniqueSchema bool `gorm:"default:true"`

	Themes   []Theme   `gorm:"many2many:tenant_themes;"`
	Billings []Billing

	TenantConfigDetail TenantConfigDetail

	// Denormalised region fields used for queries and unique index.
	RegionName       string `gorm:"type:varchar(255)"`
	RegionRootDomain string `gorm:"type:varchar(255);uniqueIndex:idx_subdomain_region"`
}
