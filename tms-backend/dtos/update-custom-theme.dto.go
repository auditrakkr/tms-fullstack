package dto

type UpdateCustomThemeDto struct {
    ID          int     `json:"id" validate:"required"`        // Required field
    Name        *string `json:"name,omitempty"`               // Optional field
    Description *string `json:"description,omitempty"`        // Optional field
    Properties  *string `json:"properties,omitempty"`         // Optional field
}