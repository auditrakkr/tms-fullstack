package dto

type UpdateTenantConfigDetailDto struct {
    WebServerProperties   *WebServerProperties       `json:"webServerProperties,omitempty"`   // Optional field
    DBProperties          *DBProperties              `json:"dbProperties,omitempty"`          // Optional field
    DBSchema              *string                    `json:"dbSchema,omitempty"`              // Optional field
    ElasticSearchProperties *ElasticSearchProperties `json:"elasticSearchProperties,omitempty"` // Optional field
    RedisProperties       *RedisProperties           `json:"redisProperties,omitempty"`       // Optional field
    RootFileSystem        *RootFileSystem            `json:"rootFileSystem,omitempty"`        // Optional field
    SMTPAuth              *SMTPAuth                  `json:"smtpAuth,omitempty"`              // Optional field
	JWTConstants          *JWTConstants              `json:"jwtConstants,omitempty"`          // Optional field
    AuthEnabled           *AuthEnabled               `json:"authEnabled,omitempty"`           // Optional field
    FBOauth2Constants     *FBOauth2Constants         `json:"fbOauth2Constants,omitempty"`     // Optional field
    GoogleOidcConstants   *GoogleOauth2Constants      `json:"googleOidcConstants,omitempty"`   // Optional field
    OtherUserOptions      *OtherUserOptions          `json:"otherUserOptions,omitempty"`      // Optional field
    SizeLimits            *SizeLimits                `json:"sizeLimits,omitempty"`            // Optional field
    Theme                 *ThemeType                     `json:"theme,omitempty"`                 // Optional field
    Logo                  *Logo                      `json:"logo,omitempty"`                  // Optional field
    Tenant                *CreateTenantDto           `json:"tenant,omitempty"`                // Optional field
    Region                *CreateRegionDto           `json:"region,omitempty"`                // Optional field
}