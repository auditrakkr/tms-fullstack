package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

type TenantAccountOfficer struct {
	gorm.Model
	TenantID uint
	UserID uint
	Tenant Tenant `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
	//  Denormalizing roles  e.g. manager, tech-support, etc. for efficiency of access for display on the client side
	Roles []global.TenantAccountOfficerRole
}