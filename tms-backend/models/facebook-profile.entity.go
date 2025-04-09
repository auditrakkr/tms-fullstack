package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Emails struct {
	Value string `json:"value"`
	Type *string `json:"type"`
}

// Scan implements the sql.Scanner interface for deserializing JSONB from the database
func (e *Emails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, e)
}

// Value implements the driver.Valuer interface for serializing to JSONB in the database
func (e Emails) ToDriverValue() (driver.Value, error) {
	return json.Marshal(e)
}

type Name struct {
	FamilyName string `json:"family_name"`
	GivenName string `json:"given_name"`
}

// Scan implements the sql.Scanner interface for deserializing JSONB from the database
func (n *Name) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, n)
}

// Value implements the driver.Valuer interface for serializing to JSONB in the database
func (n Name) ToDriverValue() (driver.Value, error) {
	return json.Marshal(n)
}


// FacebookProfile represents a user's Facebook profile information

type FacebookProfile struct {
	gorm.Model
	UserID uint `gorm:"constraint:OnDelete:CASCADE"`
	FacebookID string `gorm:"type:varchar(255);unique;index;not null"`
	DisplayName string `gorm:"type:varchar(255)"`
	Photos string `gorm:"type:varchar(255)"`
	Emails Emails `gorm:"type:jsonb"`
	Gender string `gorm:"type:varchar(255)"`
	Name Name `gorm:"type:jsonb"`
	ProfileURL string `gorm:"type:varchar(255)"`

}