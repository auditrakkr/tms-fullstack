package models

import "gorm.io/gorm"


type WebServerProperties struct {
	Host string `json:"host"`
	Port *int `json:"port"`
	Login *string `json:"login"`
	Password *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"`
	} `json:"password"`
}
type DBProperties struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port int `json:"port"`
	Username string `json:"username"`
	Password *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"` // Fixed: Added the missing backtick
	} `json:"password"`
	Database string `json:"database"`
	SSL *SSL `json:"ssl"`
}
type SSL struct {
	RejectUnauthorized bool `json:"reject_unauthorized"`  // Whether SSL is required
	CA string `json:"ca"`           // SSL mode (disable, require, verify-ca, verify-full)
	Cert *string `json:"cert,omitempty"` // Path to client certificate
	Key *string `json:"key,omitempty"`  // Path to client key
}

type ElasticSearchProperties struct {
	Node string `json:"node"`
	Username string `json:"username"`
	Password *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"`
	} `json:"password"`
	CA *string `json:"ca"` //public key for elasticsearch if using 9300 secure port. See https://www.elastic.co/guide/en/elasticsearch/reference/current/security-basic-setup-https.html for secure setup
} //for elastic search connection. This may not be different for each client. If not set, use the general one
type RedisProperties struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Password *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"`
	} `json:"password"`
	DB *int `json:"db"`
	Sentinels *string `json:"sentinels"` //supposed to be { host: string, port: number }[]
	Family int `json:"family"`
	CA *string `json:"ca"` //public key for redis if using secure port. See https://redis.io/topics/ssl for secure setup
} //for redis connection.
type RootFileSystem struct {
	Path string `json:"path"`
	Username *string `json:"username"`
	Password *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"`
	} `json:"password"`
	CA *string `json:"ca"` //public key for file system if using secure port. See https://www.openssl.org/docs/man1.1.1/man1/s_server.html for secure setup
} //root of filesystem for uploads for the tenant. Each tenant in the region should have a suffix based on tenant's uuid
type SMTPAuth struct {
	SMTPUser string `json:"smtp_user"`
	SMTPPwd *struct {
		IV *string `json:"iv"`
		Content *string `json:"content"`
	} `json:"smtp_pwd"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort int `json:"smtp_port"`
	SMTPService string `json:"smtp_service"`
	SMTPSecure string `json:"smtp_secure"`
	SMTPOauth bool `json:"smtp_oauth"`
	SMTPClientID *string `json:"smtp_client_id"`
	SMTPClientSecret *string `json:"smtp_client_secret"`
	SMTPAccessToken *string `json:"smtp_access_token"`
	SMTPRefreshToken *string `json:"smtp_refresh_token"`
	SMTPAccessUrl *string `json:"smtp_access_url"`
	SMTPPool bool `json:"smtp_pool"`	
	SMTPMaxConnections int `json:"smtp_max_connections"`
	SMTPMaxMessages int `json:"smtp_max_messages"`
}
type JWTConstants struct {
	JWTSecretKeyExpiration string `json:"jwt_secret_key_expiration"`
	JWTRefreshSecretKeyExpiration string `json:"jwt_refresh_secret_key_expiration"`
	JWTSecretKey string `json:"jwt_secret_key"`
	JWTRefreshSecret string `json:"jwt_refresh_secret"`
	JWTSecretPrivateKey string `json:"jwt_secret_private_key"`
	JWTSecretPublicKey string `json:"jwt_secret_public_key"`
	JWTSigningAlgorithm string `json:"jwt_signing_algorithm"`
}

type AuthEnabled struct {
	Google bool `json:"google"`
	Facebook bool `json:"facebook"`
	TwoFactor bool `json:"two_factor"`
}

type FBOauth2Constants struct {
	FBAppID string `json:"fb_app_id"`
	FBAppSecret string `json:"fb_app_secret"`
	CreateUserIfNotExists bool `json:"create_user_if_not_exists"`
}
type GoogleOauth2Constants struct {
	GoogleOauth2ClientOidcIssuer string `json:"google_oauth2_client_oidc_issuer"`
	GoogleApiKey string `json:"google_api_key"`
	GoogleOauth2ClientID string `json:"google_oauth2_client_id"`
	GoogleOauth2ClientSecret string `json:"google_oauth2_client_secret"`
	CreateUserIfNotExists bool `json:"create_user_if_not_exists"`
}

type OtherUserOptions struct {
	ResetPasswordMailOptionSettings_TextTemplate string `json:"reset_password_mail_option_settings_text_template"`
	ConfirmEmailMailOptionSettings_TextTemplate string `json:"confirm_email_mail_option_settings_text_template"`
	PasswordResetExpiration int `json:"password_reset_expiration"`
	EmailVerificationExpiration int `json:"email_verification_expiration"`
}
type SizeLimits struct {
	LogoFileSizeLimit int `json:"logo_file_size_limit"`
	PhotoFileSizeLimit int `json:"photo_file_size_limit"`
	GeneralFileSizeLimit int `json:"general_file_size_limit"`
}

type ThemeType struct {
	Custom bool `json:"custom"`
	Type string `json:"type"`
	RootUrl string `json:"root_url"`
}

type Logo struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
}


type TenantConfigDetail struct {
	gorm.Model
	WebServerProperties *WebServerProperties `gorm:"type:jsonb"`
	DBProperties *DBProperties `gorm:"type:jsonb"`
	DBSchema string
	ElasticSearchProperties *ElasticSearchProperties `gorm:"type:jsonb"`
	RedisProperties *RedisProperties `gorm:"type:jsonb"`
	RootFileSystem *RootFileSystem `gorm:"type:jsonb"`
	SMTPAuth *SMTPAuth `gorm:"type:jsonb"`
	JWTConstants *JWTConstants `gorm:"type:jsonb"`
	AuthEnabled *AuthEnabled `gorm:"type:jsonb"`
	FBOauth2Constants *FBOauth2Constants `gorm:"type:jsonb"`
	GoogleOauth2Constants *GoogleOauth2Constants `gorm:"type:jsonb"`
	OtherUserOptions *OtherUserOptions `gorm:"type:jsonb"`
	SizeLimits *SizeLimits `gorm:"type:jsonb"`
	Theme *ThemeType `gorm:"type:jsonb"`
	Logo *Logo `gorm:"type:jsonb"`
	TenantID uint
	RegionID uint
	Region Region
	
}
