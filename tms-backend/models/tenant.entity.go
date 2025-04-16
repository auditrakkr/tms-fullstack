package models

import (
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model
	UUID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();unique;not null"`
	Name string `gorm:"type:varchar(255);not null"`
	Subdomain string `gorm:"type:varchar(255);uniqueIndex:idx_subdomain_region;not null"`
	CustomURLSlug string `gorm:"type:varchar(255);unique"`
	Address string `gorm:"type:varchar(255)"`
	MoreInfo string `gorm:"type:text"`
	Logo string `gorm:"type:varchar(255)"`
	LogoMimeType string `gorm:"type:varchar(255)"`
	Status global.TenantStatus `gorm:"type:tenant_status;default:'active'" json:"status"`
	Active bool `gorm:"default:false"`

	PrimaryContactID uint
	PrimaryContact User

	CustomTheme CustomTheme
	TenantTeams []TenantTeam
	TenantAccountOfficers []TenantAccountOfficer

	UniqueSchema bool `gorm:"default:true"`

	Themes []Theme `gorm:"many2many:tenant_themes;"`
	Billings []Billing
	//Config details for tenant
    //Connection for this tenant
	TenantConfigDetail TenantConfigDetail
	RegionName string `gorm:"type:varchar(255)"` //denormalized region unique name called getTenantsByRegionName in tenants service

	RegionRootDomain string `gorm:"type:varchar(255);uniqueIndex:idx_subdomain_region"` //denomalized so as to set up unique index with tenant name. So tenantName.rootDomainName cannot be the repeated


}

func (t *Tenant) MapFromCreateTenantDto(dto *dto.CreateTenantDto) error {
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

    return nil
}