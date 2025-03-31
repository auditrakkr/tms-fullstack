package dto

import "github.com/auditrakkr/tms-fullstack/tms-backend/global"

// CreateTenantTeamDto represents the structure for creating a tenant team
type CreateTenantTeamDto struct {
    Tenant           *CreateTenantDto   `json:"tenant,omitempty"`           // Optional field
    User             *CreateUserDto     `json:"user,omitempty"`             // Optional field
    Roles            []global.TenantTeamRole   `json:"roles" validate:"required,dive,required"` // Required array of roles
    TenantUniqueName string             `json:"tenantUniqueName" validate:"required"`    // Required field
    TenantUniqueId   int                `json:"tenantUniqueId" validate:"required"`      // Required field
}

// CreateTenantTeamRolesDto represents the structure for roles
type CreateTenantTeamRolesDto struct {
    Roles []global.TenantTeamRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}