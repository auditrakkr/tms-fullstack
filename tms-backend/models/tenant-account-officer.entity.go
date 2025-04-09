package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

/**
 * This entity is a joiner for tenant and user (account officers). For each join, there are roles associated
 */
type TenantAccountOfficer struct {
	gorm.Model
	TenantID uint `gorm:"uniqueIndex:idx_tenant_user"`
	UserID uint `gorm:"uniqueIndex:idx_tenant_user"`
	Tenant Tenant `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
	//  Denormalizing roles  e.g. manager, tech-support, etc. for efficiency of access for display on the client side
	Roles []global.TenantAccountOfficerRole `gorm:"type:tenant_account_officer_role[]"`
}