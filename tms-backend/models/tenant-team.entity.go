package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

type TenantTeam struct {
	gorm.Model
	TenantID uint
	UserID uint
	Tenant Tenant `gorm:"constraint:OnDelete:CASCADE"`
	User User `gorm:"constraint:OnDelete:CASCADE"`
	Roles []global.TenantTeamRole
    //  Denormalizing tenant unique name for efficiency of access for display on the client side 
	TenantUniqueName string
	//  Denormalizing tenant unique ID for efficiency of access on the client side
	TenantUniqueID uint

}
