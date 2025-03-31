package models

import "gorm.io/gorm"


type Billing struct {
	gorm.Model
	Code string `gorm:"type:varchar(255);not null;unique"`
	Description string `gorm:"type:text"`
	Type string `gorm:"type:varchar(255);not null"` //could be a categorization of the billing
	TenantID uint
	Tenant Tenant
}