package dto

type UpdateRoleDto struct {
    Name        *string `json:"name,omitempty"`        // Optional field
    Description *string `json:"description,omitempty"` // Optional field
    Landlord    *bool   `json:"landlord,omitempty"`    // Optional field
}

type UpdateRoleDtos struct {
    Dtos []UpdateRoleDto `json:"dtos"` // Array of UpdateRoleDto
}