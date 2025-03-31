package dto

type CreateThemeDto struct {
    Name        string `json:"name" validate:"required"`        // Required field
    Description string `json:"description" validate:"required"` // Required field
    Properties  string `json:"properties" validate:"required"`  // Required field
}