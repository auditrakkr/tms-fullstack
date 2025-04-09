package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

/**
 * This entity is a joiner for tenant and user (team). For each join, there are roles associated
 */
type TenantTeam struct {
    gorm.Model
    TenantID uint `gorm:"uniqueIndex:idx_tenant_team_user"`
    UserID uint `gorm:"uniqueIndex:idx_tenant_team_user"`
    Tenant Tenant `gorm:"constraint:OnDelete:CASCADE"`
    User User `gorm:"constraint:OnDelete:CASCADE"`
    Roles []global.TenantTeamRole `gorm:"type:tenant_team_role[]"`
    //  Denormalizing tenant unique name for efficiency of access for display on the client side
    TenantUniqueName string
    //  Denormalizing tenant unique ID for efficiency of access on the client side
    TenantUniqueID uint
}
