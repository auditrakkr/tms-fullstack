package models

import (
	"gorm.io/gorm"
)

type Region struct {
    gorm.Model
    Name string `gorm:"type:varchar(255);not null;index;unique"`
    RootDomainName string `gorm:"type:varchar(255);not null"`
    Description *string `gorm:"type:text"`
    Country *string `gorm:"type:varchar(255)"`
    City *string `gorm:"type:varchar(255)"`
    TenantCountCapacity int `gorm:"type:int;default:5"`
    WebServerProperties *WebServerProperties `gorm:"type:jsonb"`
    DBProperties DBProperties `gorm:"type:jsonb"`
    ElasticSearchProperties *ElasticSearchProperties `gorm:"type:jsonb"`
    RedisProperties *RedisProperties `gorm:"type:jsonb"`
    RootFileSystem *RootFileSystem `gorm:"type:jsonb"`
    SMTPAuth *SMTPAuth `gorm:"type:jsonb"`
    JWTConstants *JWTConstants `gorm:"type:jsonb"`
    AuthEnabled *AuthEnabled `gorm:"type:jsonb"`
    FBOauth2Constants *FBOauth2Constants `gorm:"type:jsonb"`
    GoogleOidcConstants *GoogleOauth2Constants `gorm:"type:jsonb"`
    OtherUserOptions *OtherUserOptions `gorm:"type:jsonb"`
    SizeLimits *SizeLimits `gorm:"type:jsonb"`
    Theme *ThemeType `gorm:"type:jsonb"`
    TenantConfigDetails []TenantConfigDetail
}