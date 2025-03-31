package models

import "gorm.io/gorm"

type TailwindProperties struct {
	PrimaryBackground string `json:"primaryBackground"`
	PrimaryColor string `json:"primaryColor"`
}

type CustomTheme struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null"`
	Description	string `gorm:"type:text"`
	Properties string `gorm:"type:jsonb"`
	TailwiindConfig TailwindProperties `gorm:"type:jsonb"`

	TenantID uint
}