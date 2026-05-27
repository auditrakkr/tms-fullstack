package tenants

import (
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
)

// TenantTeamRoles is the request body for team-member role operations.
type TenantTeamRoles struct {
	Roles []global.TenantTeamRole `json:"roles" validate:"required,dive,required"`
}

// TenantAccountOfficerRoles is the request body for account-officer role operations.
type TenantAccountOfficerRoles struct {
	Roles []global.TenantAccountOfficerRole `json:"roles" validate:"required,dive,required"`
}

// TenantIdentity bundles the subset of tenant + primary-contact fields needed
// by the multi-tenancy engine to build unique prefixes and Redis keys.
type TenantIdentity struct {
	UUID          string
	SubDomainName string
	CustomURLSlug string
	RegionName    string
	Name          string
	MoreInfo      string
	Address       string
	Status        global.TenantStatus
	PrimaryContact struct {
		FirstName           string
		LastName            string
		PrimaryEmailAddress string
		Gender              global.Gender
		DateOfBirth         time.Time
	}
}
