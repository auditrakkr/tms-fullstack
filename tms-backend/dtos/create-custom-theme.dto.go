package dto

type CreateCustomThemeDto struct {
    Name        string  `json:"name" validate:"required"` // Required field
    Description *string `json:"description,omitempty"`   // Optional field
    Properties  *string `json:"properties,omitempty"`    // Optional field
}