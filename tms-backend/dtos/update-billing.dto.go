package dto

type UpdateBillingDto struct {
    Code        *string `json:"code,omitempty"`        // Optional field
    Description *string `json:"description,omitempty"` // Optional field
    Type        *string `json:"type,omitempty"`        // Optional field
}