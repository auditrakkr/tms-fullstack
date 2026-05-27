package global

import (
	"fmt"
	"net/http"
	"time"
)

// ── Role / status enums ──────────────────────────────────────────────────────

type TenantStatus string

const (
	Active    TenantStatus = "active"
	Suspended TenantStatus = "suspended"
	Owing     TenantStatus = "owing"
)

type TenantTeamRole string

const (
	A TenantTeamRole = "Admin"
	M TenantTeamRole = "Manager"
	E TenantTeamRole = "Employee"
)

type TenantAccountOfficerRole string

const (
	AOM TenantAccountOfficerRole = "Account Officer Manager"
	AOT TenantAccountOfficerRole = "Account Officer Tech-Support"
)

type LandlordRoles string

const (
	AdminLandlord      LandlordRoles = "admin_landlord"
	SuperAdminLandlord LandlordRoles = "super_admin_landlord"
	UserLandlord       LandlordRoles = "user_landlord"
)

type TenantRoles string

const (
	Admin      TenantRoles = "admin"
	SuperAdmin TenantRoles = "super_admin"
	User       TenantRoles = "user"
)

// ── Compile-time constants ────────────────────────────────────────────────────
// These are fixed by the application design and are not intended to be
// overridden via environment variables.

const (
	APP_NAME        = "TMS"
	APP_VERSION     = "1.0.0"
	APP_DESCRIPTION = "Tenant Management System"
	API_VERSION     = "v1"
	API_BASE_URL    = "/api/v1"

	USE_API_VERSION_IN_URL = true

	PASSWORD_RESET_EXPIRATION     = 24 * time.Hour // 24 hours
	EMAIL_VERIFICATION_EXPIRATION = 48 * time.Hour // 48 hours
)

// ── Runtime-configurable settings ────────────────────────────────────────────
// These vars are set by InitAppSettings (called from main after LoadConfig).
// They start with safe defaults so the app can still run without a full .env.

var (
	// Protocol used when building email verification / reset URLs.
	// Overridden by HTTP_PROTOCOL env var.
	PROTOCOL = "http"

	// Upload directory for user photos and logos.
	// Overridden by UPLOAD_DIRECTORY env var.
	UPLOAD_DIRECTORY = "uploads"

	// Maximum file sizes in bytes.
	// Overridden by PHOTO_FILE_SIZE_LIMIT / LOGO_FILE_SIZE_LIMIT env vars.
	PHOTO_FILE_SIZE_LIMIT   = int64(2 * 1024 * 1024) // 2 MB
	LOGO_FILE_SIZE_LIMIT    = int64(1 * 1024 * 1024) // 1 MB
	GENERAL_FILE_SIZE_LIMIT = int64(5 * 1024 * 1024) // 5 MB

	// Email sending behaviour.
	AUTO_SEND_CONFIRM_EMAIL = true
	AUTO_SEND_WELCOME_EMAIL = true
)

// AppSettingsConfig is the slice of config.Config that InitAppSettings needs.
// Defined as an interface so global does not import config (avoiding cycles).
type AppSettingsConfig interface {
	GetProtocol() string
	GetUploadDirectory() string
	GetPhotoFileSizeLimit() int64
	GetLogoFileSizeLimit() int64
	GetGeneralFileSizeLimit() int64
}

// InitAppSettings populates runtime vars from the application config.
// Call once from main() after config.LoadConfig().
func InitAppSettings(protocol, uploadDir string, photoLimit, logoLimit, generalLimit int64) {
	if protocol != "" {
		PROTOCOL = protocol
	}
	if uploadDir != "" {
		UPLOAD_DIRECTORY = uploadDir
	}
	if photoLimit > 0 {
		PHOTO_FILE_SIZE_LIMIT = photoLimit
	}
	if logoLimit > 0 {
		LOGO_FILE_SIZE_LIMIT = logoLimit
	}
	if generalLimit > 0 {
		GENERAL_FILE_SIZE_LIMIT = generalLimit
	}
}

// ── URL helpers ───────────────────────────────────────────────────────────────

// BuildBaseURL constructs the scheme+host from the incoming request.
// Respects X-Forwarded-Proto for reverse-proxy deployments.
func BuildBaseURL(r *http.Request) string {
	protocol := r.Header.Get("X-Forwarded-Proto")
	if protocol == "" {
		if r.TLS != nil {
			protocol = "https"
		} else {
			protocol = PROTOCOL
		}
	}
	return fmt.Sprintf("%s://%s", protocol, r.Host)
}

// APIBaseURL returns the full base URL including the /api/v1 prefix.
func APIBaseURL(r *http.Request) string {
	base := BuildBaseURL(r)
	if USE_API_VERSION_IN_URL {
		return base + API_BASE_URL
	}
	return base
}
