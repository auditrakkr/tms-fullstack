package database

import (
	"fmt"
	"log"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

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
