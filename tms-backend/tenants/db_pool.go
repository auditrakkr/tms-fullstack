package tenants

import (
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// dbPool holds open *gorm.DB instances keyed by connection name.
// A name is typically the region name for shared connections, or
// region_name + tenantUniquePrefix for tenant-specific databases.
type dbPool struct {
	conns sync.Map
}

func newDBPool() *dbPool {
	return &dbPool{}
}

// DBConnectionConfig holds the parameters needed to open a new PostgreSQL connection.
type DBConnectionConfig struct {
	Host     string
	Username string
	Password string
	Database string
	Port     int
	SSLMode  string // "disable", "require", etc.
}

// Get returns an existing connection for name, or creates and caches a new one.
func (p *dbPool) Get(name string, cfg DBConnectionConfig) (*gorm.DB, error) {
	if v, ok := p.conns.Load(name); ok {
		return v.(*gorm.DB), nil
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, sslMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open tenant DB connection %q: %w", name, err)
	}
	p.conns.Store(name, db)
	return db, nil
}
