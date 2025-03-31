package dto

import "github.com/auditrakkr/tms-fullstack/tms-backend/global"

// UpdateTenantAccountOfficerDto represents the structure for updating a tenant account officer
type UpdateTenantAccountOfficerDto struct {
    ID     *int                     `json:"id,omitempty"`     // Optional field
    Tenant *CreateTenantDto         `json:"tenant,omitempty"` // Optional field
    User   *CreateUserDto           `json:"user,omitempty"`   // Optional field
    Roles  []global.TenantAccountOfficerRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}

// UpdateTenantAccountOfficerRolesDto represents the structure for updating roles
type UpdateTenantAccountOfficerRolesDto struct {
    Roles []global.TenantAccountOfficerRole `json:"roles" validate:"required,dive,required"` // Required array of roles
}