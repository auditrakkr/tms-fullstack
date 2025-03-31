package models

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

type Theme struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null"`
	Type global.ThemeType `gorm:"type:enum('standard', 'auditrakkr');default:'standard'" json:"type"`
	Description string `gorm:"type:text"`
	Properties string `gorm:"type:jsonb"`
	Tenants []Tenant `gorm:"many2many:tenant_themes;"`
}
