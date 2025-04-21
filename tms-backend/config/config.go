package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type Config struct {
	Postgres struct{
		Host     string
		Port     int
		User     string
		Password string
		DB       string
		SSLMode  *string
	}
	// Redis configuration
	Redis struct {
		Host     string
		Port     int
		Password string
	}
	Elasticsearch    *struct {
		Node     *string
		Username *string
		Password *string
	}
	SMTP *struct {
		Host     *string
		Port     *int
		User     *string
		Password *string
	}

	JWT struct {
		Secret                   string
		SecretKeyExpiration      int
		RefreshSecret            string
		RefreshSecretKeyExpiration int
	}
}
var AppConfig *Config
var AppConfigFilePath string
var (
	GoogleOAuthConfig *oauth2.Config
	FBOAuthConfig     *oauth2.Config
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	viper.AutomaticEnv()


	// Load the configuration from a file or environment variables
	// and populate the AppConfig variable.
	// This is just a placeholder implementation.
	AppConfig = &Config{}

	// PostgreSQL configuration
	AppConfig.Postgres.Host = viper.GetString("POSTGRES_HOST")
	AppConfig.Postgres.Port = viper.GetInt("POSTGRES_PORT")
	AppConfig.Postgres.User = viper.GetString("POSTGRES_USER")
	AppConfig.Postgres.Password = viper.GetString("POSTGRES_PASSWORD")
	AppConfig.Postgres.DB = viper.GetString("POSTGRES_DB")

	// Redis configuration
	AppConfig.Redis.Host = viper.GetString("REDIS_HOST")
	AppConfig.Redis.Port = viper.GetInt("REDIS_PORT")
	AppConfig.Redis.Password = viper.GetString("REDIS_PASSWORD")

	// JWT configuration
	AppConfig.JWT.Secret = viper.GetString("SECRET_KEY")
	AppConfig.JWT.SecretKeyExpiration = viper.GetInt("SECRET_KEY_EXPIRATION")
	AppConfig.JWT.RefreshSecret = viper.GetString("REFRESH_SECRET")
	AppConfig.JWT.RefreshSecretKeyExpiration = viper.GetInt("REFRESH_SECRET_KEY_EXPIRATION")
	// Configure OAuth2 for Google and Facebook
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
		ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  viper.GetString("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	FBOAuthConfig = &oauth2.Config{
		ClientID:     viper.GetString("FACEBOOK_APP_ID"),
		ClientSecret: viper.GetString("FACEBOOK_APP_SECRET"),
		RedirectURL:  viper.GetString("FACEBOOK_REDIRECT_URL"),
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}

}