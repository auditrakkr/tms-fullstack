package dto

type CreateRoleDto struct {
    Name        string  `json:"name" validate:"required"`        // Required field
    Description *string `json:"description,omitempty"`          // Optional field
    Landlord    *bool   `json:"landlord,omitempty"`             // Optional field
}

type CreateRoleDtos struct {
    Dtos []CreateRoleDto `json:"dtos"` // Array of CreateRoleDto
}