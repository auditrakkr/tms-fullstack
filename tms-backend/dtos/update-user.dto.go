package dto

import (
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
)

// UpdateUserDto represents the structure for updating a user
type UpdateUserDto struct {
    ID                          *int      `json:"id,omitempty"`                          // Optional field
    Landlord                    *bool     `json:"landlord,omitempty"`                   // Optional field
    FirstName                   *string   `json:"firstName,omitempty"`                  // Optional field
    MiddleName                  *string   `json:"middleName,omitempty"`                 // Optional field
    LastName                    *string   `json:"lastName,omitempty"`                   // Optional field
    CommonName                  *string   `json:"commonName,omitempty"`                 // Optional field
    HomeAddress                 *string   `json:"homeAddress,omitempty"`                // Optional field
    Gender                      *global.Gender   `json:"gender,omitempty"`                     // Optional field
    DateOfBirth                 *time.Time `json:"dateOfBirth,omitempty"`               // Optional field
    Nationality                 *string   `json:"nationality,omitempty"`                // Optional field
    StateOfOrigin               *string   `json:"stateOfOrigin,omitempty"`              // Optional field
    Zip                         *string   `json:"zip,omitempty"`                        // Optional field
    Photo                       *string   `json:"photo,omitempty"`                      // Optional field
    PhotoMimeType               *string   `json:"photoMimeType,omitempty"`              // Optional field
    IsActive                    *bool     `json:"isActive,omitempty"`                   // Optional field
    PrimaryEmailAddress         *string   `json:"primaryEmailAddress,omitempty" validate:"omitempty,email"` // Optional field with email validation
    BackupEmailAddress          *string   `json:"backupEmailAddress,omitempty" validate:"omitempty,email"`  // Optional field with email validation
    Phone                       *PhoneDto `json:"phone,omitempty"`                      // Optional field
    IsPrimaryEmailAddressVerified *bool   `json:"isPrimaryEmailAddressVerified,omitempty"` // Optional field
    PasswordSalt                *string   `json:"passwordSalt,omitempty"`               // Optional field
    PasswordHash                *string   `json:"passwordHash,omitempty"`               // Optional field
    IsPasswordChangeRequired    *bool     `json:"isPasswordChangeRequired,omitempty"`   // Optional field
    RefreshTokenHash            *string   `json:"refreshTokenHash,omitempty"`           // Optional field
}