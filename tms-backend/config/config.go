package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisHost        *string
	RedisPort        *int
	RedisPassword    *string
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
}
var AppConfig *Config
var AppConfigFilePath string

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")	
	}
	
	viper.AutomaticEnv()
	

	// Load the configuration from a file or environment variables
	// and populate the AppConfig variable.
	// This is just a placeholder implementation.
	AppConfig = &Config{
		PostgresHost:     viper.GetString("POSTGRES_HOST"),
		PostgresPort:     viper.GetInt("POSTGRES_PORT"),
		PostgresUser:     viper.GetString("POSTGRES_USER"),
		PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
		PostgresDB:       viper.GetString("POSTGRES_DB"),
	}
}