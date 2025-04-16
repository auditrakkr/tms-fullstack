package database

import (
	"context"
	"fmt"
	"log"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
	Cache *redis.Client
	ctx = context.Background()
)

func ConnectDB() {
	cfg := config.AppConfig

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DB, cfg.Postgres.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Assign the db instance to the global DB variable
    DB = db

	//Connect to Redis for caching
	Cache = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB: 0, // use default DB
	})

	// Test Redis connection
    _, err = Cache.Ping(ctx).Result()
    if err != nil {
        log.Printf("Warning: Failed to connect to Redis cache: %v", err)
        log.Println("Continuing without cache functionality...")
        Cache = nil
    } else {
        log.Println("Connected to Redis cache successfully")
    }

	// Create the enum type if it doesn't exist
	err = DB.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender') THEN
				CREATE TYPE gender AS ENUM ('male', 'female');
			END IF;

			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_status') THEN
				CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'owing');
			END IF;

			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_team_role') THEN
				CREATE TYPE tenant_team_role AS ENUM ('Admin', 'Manager', 'Employee');
			END IF;

			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_account_officer_role') THEN
				CREATE TYPE tenant_account_officer_role AS ENUM ('Account Officer Manager', 'Account Officer Tech-Support');
			END IF;

			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'theme_type') THEN
				CREATE TYPE theme_type AS ENUM ('standard', 'auditrakkr');
			END IF;
		END $$;
	`).Error
	if err != nil {
		log.Fatalf("Failed to create enum type: %v", err)
	}

	// Ensure uuid-ossp extension is available
	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.Billing{},
		&models.CustomTheme{},
		&models.Region{},
		&models.Role{},
		&models.TenantAccountOfficer{},
		&models.TenantConfigDetail{},
		&models.TenantTeam{},
		&models.Theme{},
		&models.FacebookProfile{},
		&models.GoogleProfile{},
		); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}


	log.Println("Connected to PostgreSQL database successfully")
}

func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("Failed to get database connection: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
		log.Println("Database connection closed")
	}

	// Close the Redis client
	if Cache != nil {
		if err := Cache.Close(); err != nil {
			log.Fatalf("Failed to close Redis connection: %v", err)
		}
		log.Println("Redis connection closed")
	}

}

/* func SetCache(key string, value interface{}) error {
	if Cache == nil {

	}
} */