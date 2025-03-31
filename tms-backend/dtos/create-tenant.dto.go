package dto

import (
	"time"

	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
)

type CreateTenantDto struct {
	Name                 string                       `json:"name" validate:"required"`
	Address              string                       `json:"address" validate:"required"`
	MoreInfo             *string                      `json:"moreInfo,omitempty"`
	Logo                 *string                      `json:"logo,omitempty"`
	LogoMimeType         *string                      `json:"logoMimeType,omitempty"`
	DateOfRegistration   time.Time                    `json:"dateOfRegistration" validate:"required"`
	Status               *global.TenantStatus                `json:"status,omitempty"`
	PrimaryContact       *CreateUserDto               `json:"primaryContact,omitempty"`
	CustomTheme          *CreateCustomThemeDto        `json:"customTheme,omitempty"`
	TenantConfigDetail   *CreateTenantConfigDetailDto `json:"tenantConfigDetail,omitempty"`
	RegionName           string                       `json:"regionName" validate:"required"`
	RegionRootDomainName *string                      `json:"regionRootDomainName,omitempty"`
}

type CreateTenantDtos struct {
	Dtos []CreateTenantDto `json:"dtos"`
}
