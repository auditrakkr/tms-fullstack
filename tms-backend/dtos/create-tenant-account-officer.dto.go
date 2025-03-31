package dto

import "github.com/auditrakkr/tms-fullstack/tms-backend/global"

// CreateTenantAccountOfficerDto represents the structure for creating a tenant account officer
type CreateTenantAccountOfficerDto struct {
    Tenant *CreateTenantDto          `json:"tenant,omitempty"` // Optional field
    User   *CreateUserDto            `json:"user,omitempty"`   // Optional field
    Roles  []global.TenantAccountOfficerRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}

// CreateTenantAccountOfficerRolesDto represents the structure for roles
type CreateTenantAccountOfficerRolesDto struct {
    Roles []global.TenantAccountOfficerRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}