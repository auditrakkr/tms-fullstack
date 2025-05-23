package dto

type CreateRegionDto struct {
    Name                  string                     `json:"name"`
    RootDomainName        string                     `json:"rootDomainName"`
    Description           *string                    `json:"description,omitempty"`
    Country               *string                    `json:"country,omitempty"`
    City                  *string                    `json:"city,omitempty"`
    TenantCountCapacity   *int                       `json:"tenantCountCapacity,omitempty"`
    WebServerProperties   *WebServerProperties       `json:"webServerProperties,omitempty"`
    DBProperties          DBProperties               `json:"dbProperties"`
    ElasticSearchProperties ElasticSearchProperties  `json:"elasticSearchProperties"`
    RedisProperties       RedisProperties            `json:"redisProperties"`
    RootFileSystem        RootFileSystem             `json:"rootFileSystem"`
    SMTPAuth              *SMTPAuth                  `json:"smtpAuth,omitempty"`
    JWTConstants          *JWTConstants              `json:"jwtConstants,omitempty"`
    AuthEnabled           *AuthEnabled               `json:"authEnabled,omitempty"`
    FBOauth2Constants     *FBOauth2Constants         `json:"fbOauth2Constants,omitempty"`
    GoogleOidcConstants   *GoogleOauth2Constants       `json:"googleOidcConstants,omitempty"`
    OtherUserOptions      *OtherUserOptions          `json:"otherUserOptions,omitempty"`
    SizeLimits            *SizeLimits                `json:"sizeLimits,omitempty"`
    Theme                 *ThemeType                     `json:"theme,omitempty"`
    TenantConfigDetails   *[]CreateTenantConfigDetailDto `json:"tenantConfigDetails,omitempty"`
}