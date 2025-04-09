package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)


type BirthDate struct {
	Month uint `json:"month"`
	Day uint `json:"day"`
	Year *uint `json:"year"`
}


// Scan implements the sql.Scanner interface for deserializing JSONB from the database
func (b *BirthDate) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, b)
}

// Value implements the driver.Valuer interface for serializing to JSONB in the database
func (b BirthDate) Value() (driver.Value, error) {
	return json.Marshal(b)
}

type GoogleProfile struct {
	gorm.Model
	UserID uint `gorm:"constraint:OnDelete:CASCADE"`
	GoogleID string `gorm:"type:varchar(255);unique;index;not null"`
	GivenName string `gorm:"type:varchar(255)"`
	FamilyName string `gorm:"type:varchar(255)"`
	Name string `gorm:"type:varchar(255)"`
	Gender string `gorm:"type:varchar(255)"`
	BirthDate BirthDate `gorm:"type:jsonb"`
	Email string `gorm:"type:varchar(255);unique;index;not null"`
	EmailVerified bool `gorm:"index;default:false"`
	Picture string `gorm:"type:varchar(255)"`
	Profile string `gorm:"type:varchar(255)"`
	AccessToken string `gorm:"type:varchar(255)"`
	RefreshToken string `gorm:"type:varchar(255)"`
	Exp int64 `gorm:"type:bigint"`
	HD string `gorm:"type:varchar(255)"`

}