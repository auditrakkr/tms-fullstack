package dtos

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
)

// FacebookProfileDto represents a Facebook OAuth profile
type FacebookProfileDto struct {
	User        *models.User `json:"user,omitempty"`
	FacebookID  *string      `json:"facebookId,omitempty"`
	DisplayName *string      `json:"displayName,omitempty"`
	Photos      *[]struct {
		Value string `json:"value"`
	} `json:"photos,omitempty"`
	Email  *string `json:"email,omitempty"`
	Gender *string `json:"gender,omitempty"`
	Name   *struct {
		FamilyName string `json:"familyName"`
		GivenName  string `json:"givenName"`
	} `json:"name,omitempty"`
	PhotoURL *string `json:"photoUrl,omitempty"`
}
