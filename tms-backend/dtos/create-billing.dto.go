package dto

type CreateBillingDto struct {
    Code        string `json:"code" validate:"required"`        // Required field
    Description string `json:"description" validate:"required"` // Required field
    Type        string `json:"type" validate:"required"`        // Required field
}