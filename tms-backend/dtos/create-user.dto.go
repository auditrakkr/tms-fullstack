package dto

import (
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
)

// PhoneDto represents the phone field with mobile, office, and home numbers
type PhoneDto struct {
	Mobile []string `json:"mobile,omitempty"` // Optional field
	Office []string `json:"office,omitempty"` // Optional field
	Home   []string `json:"home,omitempty"`   // Optional field
}

type CreateUserDto struct {
	Landlord                      *bool         `json:"landlord,omitempty"`                                      // Optional field
	FirstName                     string        `json:"firstName" validate:"required"`                           // Required field
	MiddleName                    *string       `json:"middleName,omitempty"`                                    // Optional field
	LastName                      string        `json:"lastName" validate:"required"`                            // Required field
	CommonName                    *string       `json:"commonName,omitempty"`                                    // Optional field
	HomeAddress                   *string       `json:"homeAddress,omitempty"`                                   // Optional field
	Gender                        global.Gender `json:"gender" validate:"required"`                              // Required field
	DateOfBirth                   time.Time     `json:"dateOfBirth" validate:"required"`                         // Required field
	Nationality                   *string       `json:"nationality,omitempty"`                                   // Optional field
	StateOfOrigin                 *string       `json:"stateOfOrigin,omitempty"`                                 // Optional field
	Zip                           *string       `json:"zip,omitempty"`                                           // Optional field
	Photo                         *string       `json:"photo,omitempty"`                                         // Optional field
	PhotoMimeType                 *string       `json:"photoMimeType,omitempty"`                                 // Optional field
	IsActive                      *bool         `json:"isActive,omitempty"`                                      // Optional field
	PrimaryEmailAddress           string        `json:"primaryEmailAddress" validate:"required,email"`           // Required field with email validation
	BackupEmailAddress            *string       `json:"backupEmailAddress,omitempty" validate:"omitempty,email"` // Optional field with email validation
	Phone                         *PhoneDto     `json:"phone,omitempty"`                                         // Optional field
	IsPrimaryEmailAddressVerified *bool         `json:"isPrimaryEmailAddressVerified,omitempty"`                 // Optional field
	PasswordSalt                  *string       `json:"passwordSalt,omitempty"`                                  // Optional field
	PasswordHash                  string        `json:"passwordHash" validate:"required"`                        // Required field
	IsPasswordChangeRequired      *bool         `json:"isPasswordChangeRequired,omitempty"`                      // Optional field
}

// CreateUserDtos represents a collection of CreateUserDto
type CreateUserDtos struct {
	Dtos []CreateUserDto `json:"dtos"`
}
