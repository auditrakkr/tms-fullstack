package dto

import "github.com/auditrakkr/tms-fullstack/tms-backend/global"


type UpdateTenantDto struct {
    Name                    *string                    `json:"name,omitempty"`                    // Optional field
    Address                 *string                    `json:"address,omitempty"`                 // Optional field
    MoreInfo                *string                    `json:"moreInfo,omitempty"`                // Optional field
    Logo                    *string                    `json:"logo,omitempty"`                    // Optional field
    LogoMimeType            *string                    `json:"logoMimeType,omitempty"`            // Optional field
    DateOfRegistration      *string                    `json:"dateOfRegistration,omitempty"`      // Optional field
    Status                  *global.TenantStatus              `json:"status,omitempty"`                  // Optional field
    PrimaryContact          *CreateUserDto             `json:"primaryContact,omitempty"`          // Optional field
    CustomTheme             *CreateCustomThemeDto      `json:"customTheme,omitempty"`             // Optional field
    TenantConfigDetail      *CreateTenantConfigDetailDto `json:"tenantConfigDetail,omitempty"`    // Optional field
    RegionName              *string                    `json:"regionName,omitempty"`              // Optional field
    RegionRootDomainName    *string                    `json:"regionRootDomainName,omitempty"`    // Optional field
}