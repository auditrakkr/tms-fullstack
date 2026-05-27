package config

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

// Config holds every runtime-configurable value for the application.
// All fields are populated by LoadConfig() from environment variables (or a
// .env file). No field should ever be hard-coded elsewhere in the codebase.
type Config struct {
	// ── Application ────────────────────────────────────────────
	App struct {
		Port        string // APP_PORT (default "8080")
		Protocol    string // HTTP_PROTOCOL (default "http")
		RootHTTPURL string // APP_ROOT_HTTP_URL
	}

	// ── PostgreSQL ──────────────────────────────────────────────
	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		DB       string
	}

	// ── Redis ───────────────────────────────────────────────────
	Redis struct {
		Host     string
		Port     int
		Password string
	}

	// ── Elasticsearch ───────────────────────────────────────────
	Elasticsearch struct {
		Node     string // ES_NODE (e.g. http://localhost:9200)
		Username string // ES_USERNAME
		Password string // ES_PASSWORD
	}

	// ── SMTP ────────────────────────────────────────────────────
	SMTP struct {
		Host     string // SMTP_HOST
		Port     int    // SMTP_PORT
		User     string // SMTP_USER
		Password string // SMTP_PASSWORD
		Secure   bool   // SMTP_SECURE
	}

	// ── JWT ─────────────────────────────────────────────────────
	JWT struct {
		Secret                     string // SECRET_KEY
		SecretKeyExpiration        int    // SECRET_KEY_EXPIRATION (seconds)
		RefreshSecret              string // REFRESH_SECRET
		RefreshSecretKeyExpiration int    // REFRESH_SECRET_KEY_EXPIRATION (seconds)
	}

	// ── Encryption ──────────────────────────────────────────────
	// Separate from JWT secret — used to encrypt region credentials at rest.
	CryptoKey string // SECRET_KEY_FOR_CRYPTO_ENCRYPTION (must be 32 bytes when padded)

	// ── Google OAuth2 / OIDC ────────────────────────────────────
	Google struct {
		APIKey       string // GOOGLE_API_KEY
		OIDCIssuer   string // GOOGLE_OAUTH2_CLIENT_OIDC_ISSUER
		ClientID     string // GOOGLE_OAUTH2_CLIENT_ID
		ClientSecret string // GOOGLE_OAUTH2_CLIENT_SECRET
		RedirectURI  string // GOOGLE_OAUTH2_REDIRECT_URI
	}

	// ── Facebook OAuth2 ─────────────────────────────────────────
	Facebook struct {
		AppID       string // FACEBOOK_APP_ID
		AppSecret   string // FACEBOOK_APP_SECRET
		CallbackURL string // CALLBACK_URL
	}

	// ── Feature toggles ─────────────────────────────────────────
	Features struct {
		EnableGoogle    bool // DEFAULT_GOOGLE
		EnableFacebook  bool // DEFAULT_FACEBOOK
		EnableTwoFactor bool // DEFAULT_TWO_FACTOR
	}

	// ── File uploads ────────────────────────────────────────────
	Uploads struct {
		Directory            string // UPLOAD_DIRECTORY
		PhotoFileSizeLimit   int64  // PHOTO_FILE_SIZE_LIMIT  (bytes)
		LogoFileSizeLimit    int64  // LOGO_FILE_SIZE_LIMIT   (bytes)
		GeneralFileSizeLimit int64  // GENERAL_FILE_SIZE_LIMIT (bytes)
	}

	// ── Default region (used for initial seeding) ────────────────
	DefaultRegion struct {
		Name              string // DEFAULT_REGION_NAME
		RootDomainName    string // DEFAULT_REGION_ROOT_DOMAIN_NAME
		CountCapacity     int    // DEFAULT_REGION_TENANT_COUNT_CAPACITY
		WebServerHost     string // DEFAULT_REGION_WEBSERVER_HOST
		WebServerPort     int    // DEFAULT_REGION_WEBSERVER_PORT
		WebServerLogin    string // DEFAULT_REGION_WEBSERVER_LOGIN
		WebServerPassword string // DEFAULT_REGION_WEBSERVER_PASSWORD
		RootFilesystemPath string // DEFAULT_ROOT_FILESYSTEM_PATH
		WebServerDomain   string // DEFAULT_REGION_WEBSERVER_DOMAIN
	}

	// ── Default mail templates ───────────────────────────────────
	Mail struct {
		FromAddress              string // defaults to noreply@auditrakkr.com
		ResetPasswordSubject     string
		ResetPasswordTemplate    string // DEFAULT_RESET_PASSWORD_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE1+LINE2+LINE3
		ConfirmEmailSubject      string
		ConfirmEmailTemplate     string // DEFAULT_CONFIRM_EMAIL_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE1+LINE2
	}
}

var AppConfig *Config

// GoogleOAuthConfig and FBOAuthConfig are the oauth2.Config instances built
// from the loaded Config; used by auth.AuthService.
var (
	GoogleOAuthConfig *oauth2.Config
	FBOAuthConfig     *oauth2.Config
)

// LoadConfig reads all environment variables (via a .env file if present)
// into AppConfig. It must be called once at startup before any other
// package uses AppConfig.
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	viper.AutomaticEnv()

	AppConfig = &Config{}

	// ── Application ──────────────────────────────────────────────
	AppConfig.App.Port = viper.GetString("APP_PORT")
	if AppConfig.App.Port == "" {
		AppConfig.App.Port = "8080"
	}
	AppConfig.App.Protocol = viper.GetString("HTTP_PROTOCOL")
	if AppConfig.App.Protocol == "" {
		AppConfig.App.Protocol = "http"
	}
	AppConfig.App.RootHTTPURL = viper.GetString("APP_ROOT_HTTP_URL")

	// ── PostgreSQL ───────────────────────────────────────────────
	AppConfig.Postgres.Host = viper.GetString("POSTGRES_HOST")
	AppConfig.Postgres.Port = viper.GetInt("POSTGRES_PORT")
	AppConfig.Postgres.User = viper.GetString("POSTGRES_USER")
	AppConfig.Postgres.Password = viper.GetString("POSTGRES_PASSWORD")
	AppConfig.Postgres.DB = viper.GetString("POSTGRES_DB")

	// ── Redis ────────────────────────────────────────────────────
	AppConfig.Redis.Host = viper.GetString("REDIS_HOST")
	AppConfig.Redis.Port = viper.GetInt("REDIS_PORT")
	AppConfig.Redis.Password = viper.GetString("REDIS_PASSWORD")

	// ── Elasticsearch ────────────────────────────────────────────
	AppConfig.Elasticsearch.Node = viper.GetString("ES_NODE")
	AppConfig.Elasticsearch.Username = viper.GetString("ES_USERNAME")
	AppConfig.Elasticsearch.Password = viper.GetString("ES_PASSWORD")

	// ── SMTP ─────────────────────────────────────────────────────
	AppConfig.SMTP.Host = viper.GetString("SMTP_HOST")
	AppConfig.SMTP.Port = viper.GetInt("SMTP_PORT")
	if AppConfig.SMTP.Port == 0 {
		AppConfig.SMTP.Port = 587
	}
	AppConfig.SMTP.User = viper.GetString("SMTP_USER")
	AppConfig.SMTP.Password = viper.GetString("SMTP_PASSWORD")
	AppConfig.SMTP.Secure = viper.GetBool("SMTP_SECURE")

	// ── JWT ──────────────────────────────────────────────────────
	AppConfig.JWT.Secret = viper.GetString("SECRET_KEY")
	AppConfig.JWT.SecretKeyExpiration = viper.GetInt("SECRET_KEY_EXPIRATION")
	AppConfig.JWT.RefreshSecret = viper.GetString("REFRESH_SECRET")
	AppConfig.JWT.RefreshSecretKeyExpiration = viper.GetInt("REFRESH_SECRET_KEY_EXPIRATION")

	// ── Encryption ───────────────────────────────────────────────
	AppConfig.CryptoKey = viper.GetString("SECRET_KEY_FOR_CRYPTO_ENCRYPTION")

	// ── Google ───────────────────────────────────────────────────
	AppConfig.Google.APIKey = viper.GetString("GOOGLE_API_KEY")
	AppConfig.Google.OIDCIssuer = viper.GetString("GOOGLE_OAUTH2_CLIENT_OIDC_ISSUER")
	AppConfig.Google.ClientID = viper.GetString("GOOGLE_OAUTH2_CLIENT_ID")
	AppConfig.Google.ClientSecret = viper.GetString("GOOGLE_OAUTH2_CLIENT_SECRET")
	AppConfig.Google.RedirectURI = viper.GetString("GOOGLE_OAUTH2_REDIRECT_URI")

	// ── Facebook ─────────────────────────────────────────────────
	AppConfig.Facebook.AppID = viper.GetString("FACEBOOK_APP_ID")
	AppConfig.Facebook.AppSecret = viper.GetString("FACEBOOK_APP_SECRET")
	AppConfig.Facebook.CallbackURL = viper.GetString("CALLBACK_URL")

	// ── Feature toggles ──────────────────────────────────────────
	AppConfig.Features.EnableGoogle = viper.GetBool("DEFAULT_GOOGLE")
	AppConfig.Features.EnableFacebook = viper.GetBool("DEFAULT_FACEBOOK")
	AppConfig.Features.EnableTwoFactor = viper.GetBool("DEFAULT_TWO_FACTOR")

	// ── Uploads ──────────────────────────────────────────────────
	AppConfig.Uploads.Directory = viper.GetString("UPLOAD_DIRECTORY")
	if AppConfig.Uploads.Directory == "" {
		AppConfig.Uploads.Directory = "uploads"
	}
	AppConfig.Uploads.PhotoFileSizeLimit = viper.GetInt64("PHOTO_FILE_SIZE_LIMIT")
	if AppConfig.Uploads.PhotoFileSizeLimit == 0 {
		AppConfig.Uploads.PhotoFileSizeLimit = 2 * 1024 * 1024 // 2 MB
	}
	AppConfig.Uploads.LogoFileSizeLimit = viper.GetInt64("LOGO_FILE_SIZE_LIMIT")
	if AppConfig.Uploads.LogoFileSizeLimit == 0 {
		AppConfig.Uploads.LogoFileSizeLimit = 1 * 1024 * 1024 // 1 MB
	}
	AppConfig.Uploads.GeneralFileSizeLimit = viper.GetInt64("GENERAL_FILE_SIZE_LIMIT")
	if AppConfig.Uploads.GeneralFileSizeLimit == 0 {
		AppConfig.Uploads.GeneralFileSizeLimit = 5 * 1024 * 1024 // 5 MB
	}

	// ── Default Region ───────────────────────────────────────────
	AppConfig.DefaultRegion.Name = viper.GetString("DEFAULT_REGION_NAME")
	AppConfig.DefaultRegion.RootDomainName = viper.GetString("DEFAULT_REGION_ROOT_DOMAIN_NAME")
	AppConfig.DefaultRegion.CountCapacity = viper.GetInt("DEFAULT_REGION_TENANT_COUNT_CAPACITY")
	if AppConfig.DefaultRegion.CountCapacity == 0 {
		AppConfig.DefaultRegion.CountCapacity = 100
	}
	AppConfig.DefaultRegion.WebServerHost = viper.GetString("DEFAULT_REGION_WEBSERVER_HOST")
	AppConfig.DefaultRegion.WebServerPort = viper.GetInt("DEFAULT_REGION_WEBSERVER_PORT")
	if AppConfig.DefaultRegion.WebServerPort == 0 {
		AppConfig.DefaultRegion.WebServerPort = 80
	}
	AppConfig.DefaultRegion.WebServerLogin = viper.GetString("DEFAULT_REGION_WEBSERVER_LOGIN")
	AppConfig.DefaultRegion.WebServerPassword = viper.GetString("DEFAULT_REGION_WEBSERVER_PASSWORD")
	AppConfig.DefaultRegion.RootFilesystemPath = viper.GetString("DEFAULT_ROOT_FILESYSTEM_PATH")
	if AppConfig.DefaultRegion.RootFilesystemPath == "" {
		AppConfig.DefaultRegion.RootFilesystemPath = "/var/www/"
	}
	AppConfig.DefaultRegion.WebServerDomain = viper.GetString("DEFAULT_REGION_WEBSERVER_DOMAIN")

	// ── Mail templates ───────────────────────────────────────────
	AppConfig.Mail.FromAddress = viper.GetString("MAIL_FROM_ADDRESS")
	if AppConfig.Mail.FromAddress == "" {
		AppConfig.Mail.FromAddress = "noreply@auditrakkr.com"
	}
	AppConfig.Mail.ResetPasswordSubject = "Password Reset Request"
	AppConfig.Mail.ConfirmEmailSubject = "Email Verification"

	// Build templates from multi-line env vars (join non-empty lines).
	resetLine1 := viper.GetString("DEFAULT_RESET_PASSWORD_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE1")
	resetLine2 := viper.GetString("DEFAULT_RESET_PASSWORD_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE2")
	resetLine3 := viper.GetString("DEFAULT_RESET_PASSWORD_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE3")
	if resetLine1 != "" {
		AppConfig.Mail.ResetPasswordTemplate = joinLines(resetLine1, resetLine2, resetLine3)
	} else {
		AppConfig.Mail.ResetPasswordTemplate = "Reset your password by clicking this link: {url}"
	}

	confirmLine1 := viper.GetString("DEFAULT_CONFIRM_EMAIL_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE1")
	confirmLine2 := viper.GetString("DEFAULT_CONFIRM_EMAIL_MAIL_OPTION_SETTINGS_TEXT_TEMPLATE_LINE2")
	if confirmLine1 != "" {
		AppConfig.Mail.ConfirmEmailTemplate = joinLines(confirmLine1, confirmLine2)
	} else {
		AppConfig.Mail.ConfirmEmailTemplate = "Please verify your email by clicking this link: {url}"
	}

	// ── Build oauth2.Config instances ────────────────────────────
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     AppConfig.Google.ClientID,
		ClientSecret: AppConfig.Google.ClientSecret,
		RedirectURL:  AppConfig.Google.RedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	FBOAuthConfig = &oauth2.Config{
		ClientID:     AppConfig.Facebook.AppID,
		ClientSecret: AppConfig.Facebook.AppSecret,
		RedirectURL:  AppConfig.Facebook.CallbackURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

func joinLines(lines ...string) string {
	result := ""
	for _, l := range lines {
		if l != "" {
			if result != "" {
				result += " "
			}
			result += l
		}
	}
	return result
}
