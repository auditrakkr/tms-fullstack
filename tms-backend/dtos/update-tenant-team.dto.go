package dto

import "github.com/auditrakkr/tms-fullstack/tms-backend/global"

// UpdateTenantTeamDto represents the structure for updating a tenant team
type UpdateTenantTeamDto struct {
    ID               *int                `json:"id,omitempty"`               // Optional field
    Tenant           *CreateTenantDto    `json:"tenant,omitempty"`           // Optional field
    User             *CreateUserDto      `json:"user,omitempty"`             // Optional field
    Roles            []global.TenantTeamRole    `json:"roles,omitempty"`            // Optional array of roles
    TenantUniqueName string              `json:"tenantUniqueName" validate:"required"` // Required field
    TenantUniqueId   int                 `json:"tenantUniqueId" validate:"required"`   // Required field
}

// UpdateTenantTeamRolesDto represents the structure for updating roles
type UpdateTenantTeamRolesDto struct {
    Roles []global.TenantTeamRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}