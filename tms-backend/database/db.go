package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
	Cache *RedisCache
)

func ConnectDB() {
	cfg := config.AppConfig

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password,
		cfg.Postgres.DB, cfg.Postgres.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	DB = db

	// Redis — non-fatal if unavailable.
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	if _, err := Redis.Ping(context.Background()).Result(); err != nil {
		slog.Warn("redis unavailable — caching disabled", "error", err)
		Redis = nil
		Cache = nil
	} else {
		slog.Info("connected to Redis")
		Cache = NewRedisCache(Redis)
	}

	// Create PostgreSQL enum types if they don't exist.
	if err := DB.Exec(`
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
	`).Error; err != nil {
		slog.Error("failed to create enum types", "error", err)
		os.Exit(1)
	}

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
		slog.Error("auto-migrate failed", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to PostgreSQL and migrations applied")
}

func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
		slog.Info("PostgreSQL connection closed")
	}
	if Redis != nil {
		Redis.Close()
		slog.Info("Redis connection closed")
		Cache = nil
	}
}
