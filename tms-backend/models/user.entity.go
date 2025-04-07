package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"gorm.io/gorm"
)

type Phone struct {
	Mobile []string `json:"mobile"`
	Home   []string `json:"home"`
	Office []string `json:"office"`
}

// Scan implements the sql.Scanner interface for deserializing JSONB from the database
func (p *Phone) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, p)
}

// Value implements the driver.Valuer interface for serializing to JSONB in the database
func (p Phone) Value() (driver.Value, error) {
	return json.Marshal(p)
}

type User struct {
	gorm.Model
	Landlord bool `gorm:"default:false"`
	FirstName string `gorm:"type:varchar(255);not null"`
	MiddleName string `gorm:"type:varchar(255)"`
	LastName string `gorm:"type:varchar(255);not null"`
	CommonName string `gorm:"type:varchar(255)"`
	HomeAddress string
	Gender global.Gender `gorm:"type:gender"`
	DateOfBirth time.Time `gorm:"type:date"`
	Nationality string
	StateOfOrigin string
	ZipCode string

	/* Photos Upload
	Plan for photo and other files to be served.
	Have a global root URL for such files e.g. '/uploads-to-stream'. Could be associated with a network storage
	*/

	Photo string
	PhotoMimeType string
	// PhotoURL string `gorm:"type:varchar(255)"`
	// PhotoThumbURL string `gorm:"type:varchar(255)"`
	// PhotoMediumURL string `gorm:"type:varchar(255)"`
	// PhotoLargeURL string `gorm:"type:varchar(255)"`
	// PhotoOriginalURL string `gorm:"type:varchar(255)"`
	IsActive bool `gorm:"default:true"`

	PrimaryEmailAddress string `gorm:"unique;not null"`
	BackupEmailAddress string `gorm:"unique"`

	PhoneNumbers Phone `gorm:"type:jsonb"`
	IsPrimaryEmailVerified bool `gorm:"default:false"`
	IsBackupEmailVerified bool `gorm:"default:false"`
	PasswordSalt string `gorm:"type:varchar(255)"`
	PasswordHash string `gorm:"type:varchar(255)"`
	IsPasswordChangeRequired bool `gorm:"default:false"`
	ResetPasswordToken string `gorm:"type:varchar(255)"`
	ResetPasswordExpiration time.Time
	PrimaryEmailVerificationToken string `gorm:"type:varchar(255)"`
	BackupEmailVerificationToken string `gorm:"type:varchar(255)"`
	EmailVerificationTokenExpiration time.Time

	// Incorporating OTP possibly for 2FA
	OTPEnabled bool `gorm:"default:false;not null"`
	OTPSecret string

	//Todo: Incorporate the user's role in the system


	PrimaryContactForTenants []Tenant `gorm:"foreignKey:PrimaryContactID"`

	TenantTeamMemberships []TenantTeam `gorm:"foreignKey:UserID"`
}
