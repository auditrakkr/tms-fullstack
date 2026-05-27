package main

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/auth"
	authdtos "github.com/auditrakkr/tms-fullstack/tms-backend/auth/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/auditrakkr/tms-fullstack/tms-backend/regions"
	"github.com/auditrakkr/tms-fullstack/tms-backend/roles"
	tenantconfigdetails "github.com/auditrakkr/tms-fullstack/tms-backend/tenant-config-details"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants/billings"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants/themes"
	"github.com/auditrakkr/tms-fullstack/tms-backend/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes is the DI root for all HTTP routes.
func SetupRoutes(router *gin.Engine, db *gorm.DB, cache *database.RedisCache) {
	// ── Global middleware ──────────────────────────────────────
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	// ── Services ───────────────────────────────────────────────
	cfg := config.AppConfig
	regionSvc := regions.NewRegionService(db, cache)
	userSvc := users.NewUserService(db, cfg.Elasticsearch.Node)
	userCtrl := users.NewUserController(userSvc)
	authSvc := auth.NewAuthService(userSvc, authdtos.GoogleProfileDto{}, authdtos.FacebookProfileDto{})
	authHandler := auth.NewAuthHandler(authSvc)
	tenantSvc := tenants.NewTenantService(db, regionSvc)
	tenantCtrl := tenants.NewTenantController(tenantSvc)

	// ── SSR routes (no /api/v1 prefix) ────────────────────────
	router.GET("/users/reset-password/:token", userCtrl.ShowResetPasswordForm)
	router.POST("/users/reset-password/:token", userCtrl.ProcessResetPassword)
	router.GET("/users/confirm-primary-email/:token", userCtrl.ConfirmPrimaryEmail)
	router.GET("/users/confirm-backup-email/:token", userCtrl.ConfirmBackupEmail)

	// ── API v1 ─────────────────────────────────────────────────
	api := router.Group("/api/v1")

	// Auth (public)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		if cfg.Features.EnableGoogle {
			authGroup.GET("/google", authHandler.GoogleRedirect)
			authGroup.GET("/google/callback", authHandler.GoogleCallback)
		}
		if cfg.Features.EnableFacebook {
			authGroup.GET("/facebook", authHandler.FacebookRedirect)
			authGroup.GET("/facebook/callback", authHandler.FacebookCallback)
		}
	}
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.JWTAuth())
	{
		authProtected.POST("/logout", authHandler.Logout)
	}

	// ── Protected API groups ───────────────────────────────────
	protected := api.Group("/")
	protected.Use(middleware.JWTAuth())

	// ── TENANTS ────────────────────────────────────────────────
	tenantGroup := protected.Group("/tenants")
	{
		// Basic CRUD
		tenantGroup.GET("/", tenantCtrl.GetAllTenants)
		tenantGroup.GET("/:id", tenantCtrl.FindOne)
		tenantGroup.GET("/get-active-tenants-in-region/:regionName", tenantCtrl.FindActiveTenantsByRegionName)
		tenantGroup.GET("/by-active/:active", tenantCtrl.FindByActive)
		tenantGroup.GET("/by-unique-name", tenantCtrl.FindByUniqueName)
		tenantGroup.GET("/check-domain", tenantCtrl.CheckDomainExists)
		tenantGroup.GET("/:id/account-officers", tenantCtrl.GetAccountOfficersByTenant)
		tenantGroup.GET("/:id/identities", tenantCtrl.GetTenantIdentities)
		tenantGroup.POST("/", tenantCtrl.CreateTenant)
		tenantGroup.POST("/bulk", tenantCtrl.InsertTenants)
		tenantGroup.PATCH("/:id", tenantCtrl.UpdateTenant)
		tenantGroup.DELETE("/:id", tenantCtrl.DeleteTenant)

		// Logo
		tenantGroup.POST("/:id/logo", tenantCtrl.SetTenantLogo)
		tenantGroup.GET("/:id/logo", tenantCtrl.GetTenantLogo)

		// Primary contact lifecycle
		tenantGroup.POST("/:id/primary-contact", tenantCtrl.CreateAndSetPrimaryContact)
		tenantGroup.PUT("/:id/primary-contact/:userId", tenantCtrl.SetPrimaryContactByID)
		tenantGroup.DELETE("/:id/primary-contact", tenantCtrl.UnsetPrimaryContactByID)

		// Team members
		tenantGroup.POST("/:id/team-members", tenantCtrl.CreateAndSetTeamMember)
		tenantGroup.PATCH("/:id/team-members/:userId", tenantCtrl.SetTeamMemberByID)
		tenantGroup.PUT("/:id/team-members/:userId", tenantCtrl.UpdateTeamMemberByID)
		tenantGroup.DELETE("/:id/team-members/:userId", tenantCtrl.DeleteTeamMemberByID)

		// Account officers
		tenantGroup.POST("/:id/account-officers", tenantCtrl.CreateAndSetTenantAccountOfficer)
		tenantGroup.PATCH("/:id/account-officers/:userId", tenantCtrl.SetTenantAccountOfficerByID)
		tenantGroup.PUT("/:id/account-officers/:userId", tenantCtrl.UpdateTenantAccountOfficerByID)
		tenantGroup.DELETE("/:id/account-officers/:userId", tenantCtrl.DeleteTenantAccountOfficerByID)

		// Custom theme
		tenantGroup.POST("/:id/custom-theme", tenantCtrl.CreateAndSetCustomTheme)
		tenantGroup.PUT("/:id/custom-theme/:themeId", tenantCtrl.SetCustomThemeByID)
		tenantGroup.DELETE("/:id/custom-theme", tenantCtrl.UnsetCustomThemeByID)

		// Theme associations
		tenantGroup.PATCH("/:id/themes/:themeId", tenantCtrl.AddThemeByID)
		tenantGroup.DELETE("/:id/themes/:themeId", tenantCtrl.RemoveThemeByID)

		// Billing associations
		tenantGroup.POST("/:id/billings", tenantCtrl.CreateAndAddBilling)
		tenantGroup.PATCH("/:id/billings/:billingId", tenantCtrl.AddBillingByID)
		tenantGroup.DELETE("/:id/billings/:billingId", tenantCtrl.RemoveBillingByID)

		// Tenant config detail links
		tenantGroup.POST("/:id/tenant-config-detail", tenantCtrl.CreateAndSetTenantConfigDetail) // Phase 2 provisioner
		tenantGroup.PUT("/:id/tenant-config-detail/:configDetailId", tenantCtrl.SetTenantConfigDetailByID)
		tenantGroup.DELETE("/:id/tenant-config-detail", tenantCtrl.UnsetTenantConfigDetailByID)
	}

	// ── ADMIN (Phase 2 Redis sync) ─────────────────────────────
	adminGroup := protected.Group("/admin")
	{
		adminGroup.POST("/regions/:name/sync-redis", tenantCtrl.SyncRegionRedis)
		adminGroup.POST("/tenants/:id/sync-redis", tenantCtrl.SyncTenantRedis)
		adminGroup.DELETE("/tenants/:id/sync-redis", tenantCtrl.UnsyncTenantRedis)
		adminGroup.PATCH("/tenants/:id/status", tenantCtrl.SetTenantStatus)
	}

	// ── THEMES & BILLINGS (global) ──────────────────────────────
	themeSvc := themes.NewThemeService(db)
	themeCtrl := themes.NewThemeController(themeSvc)
	themeGroup := protected.Group("/tenants/themes")
	{
		themeGroup.GET("/", themeCtrl.FindAll)
		themeGroup.POST("/", themeCtrl.CreateTheme)
	}

	billingSvc := billings.NewBillingService(db)
	billingCtrl := billings.NewBillingController(billingSvc)
	billingGroup := protected.Group("/tenants/billings")
	{
		billingGroup.GET("/", billingCtrl.FindAll)
		billingGroup.POST("/", billingCtrl.CreateBilling)
	}

	// ── USERS ──────────────────────────────────────────────────
	userCtrlAPI := users.NewUserController(userSvc)
	userGroup := protected.Group("/users")
	{
		userGroup.GET("/", userCtrlAPI.GetAllUsers)
		userGroup.GET("/:id", userCtrlAPI.FindOne)
		userGroup.GET("/:id/photo", userCtrlAPI.GetUserPhoto)
		userGroup.POST("/", userCtrlAPI.CreateUser)
		userGroup.POST("/bulk", userCtrlAPI.InsertUsers)
		userGroup.POST("/:id/photo", userCtrlAPI.SetUserPhoto)
		userGroup.PATCH("/:id", userCtrlAPI.UpdateUser)
		userGroup.DELETE("/:id", userCtrlAPI.DeleteUser)

		// User-side primary contact
		userGroup.POST("/:id/primary-contact-tenants", userCtrlAPI.CreateAndSetTenantForPrimaryContact)
		userGroup.PUT("/:id/primary-contact-tenants/:tenantId", userCtrlAPI.SetAsPrimaryContactByTenantID)
		userGroup.DELETE("/:id/primary-contact-tenants/:tenantId", userCtrlAPI.RemoveAsPrimaryContactByTenantID)

		// User-side team memberships
		userGroup.POST("/:id/tenant-team-memberships", userCtrlAPI.CreateTenantAndSetTeamMembership)
		userGroup.PATCH("/:id/tenant-team-memberships/:tenantId", userCtrlAPI.SetTeamMembershipByID)
		userGroup.PUT("/:id/tenant-team-memberships/:tenantId", userCtrlAPI.UpdateTeamMembershipRolesByID)
		userGroup.DELETE("/:id/tenant-team-memberships/:tenantId", userCtrlAPI.DeleteTeamMembershipByID)

		// User-side account officer assignments
		userGroup.POST("/:id/account-officer-assignments", userCtrlAPI.CreateAndSetAccountOfficerForWhichTenant)
		userGroup.PATCH("/:id/account-officer-assignments/:tenantId", userCtrlAPI.SetTenantAccountOfficerForWhichTenantByID)
		userGroup.PUT("/:id/account-officer-assignments/:tenantId", userCtrlAPI.UpdateTenantAccountOfficerForWhichTenantByID)
		userGroup.DELETE("/:id/account-officer-assignments/:tenantId", userCtrlAPI.DeleteTenantAccountOfficerForWhichTenantByID)
	}

	// ── REGIONS ────────────────────────────────────────────────
	regionCtrl := regions.NewRegionController(regionSvc)
	regionGroup := protected.Group("/regions")
	{
		regionGroup.GET("/", regionCtrl.GetAllRegions)
		regionGroup.GET("/:regionId", regionCtrl.FindOne)
		regionGroup.GET("/by-name/:name", regionCtrl.FindByName)
		regionGroup.GET("/get-tenant-assignable-regions-info", regionCtrl.GetTenantAssignableRegionsInfo)
		regionGroup.POST("/", regionCtrl.CreateRegion)
		regionGroup.POST("/insert", regionCtrl.InsertRegions)
		regionGroup.PATCH("/:regionId", regionCtrl.UpdateRegion)
		regionGroup.PATCH("/:regionId/tenant-config-detail/:tenantConfigDetailId", regionCtrl.AddTenantConfigDetailById)
		regionGroup.PATCH("/:regionId/tenant-config-details", regionCtrl.AddTenantConfigDetailsById)
		regionGroup.PUT("/", regionCtrl.SaveRegion)
		regionGroup.DELETE("/:regionId", regionCtrl.DeleteRegion)
		regionGroup.DELETE("/:regionId/tenant-config-detail/:tenantConfigDetailId", regionCtrl.RemoveTenantConfigDetailById)
		regionGroup.DELETE("/:regionId/tenant-config-details", regionCtrl.RemoveTenantConfigDetailsById)
	}

	// ── TENANT CONFIG DETAILS ──────────────────────────────────
	tcdSvc := tenantconfigdetails.NewTenantConfigDetailsService(db)
	tcdCtrl := tenantconfigdetails.NewTenantConfigDetailsController(tcdSvc)
	tcdGroup := protected.Group("/tenant-config-details")
	{
		tcdGroup.GET("/", tcdCtrl.GetAllTenantConfigDetails)
		tcdGroup.GET("/:id", tcdCtrl.FindOne)
		tcdGroup.POST("/", tcdCtrl.CreateTenantConfigDetail)
		tcdGroup.PATCH("/:id", tcdCtrl.Update)
		tcdGroup.DELETE("/:id", tcdCtrl.Delete)
	}

	// ── ROLES ──────────────────────────────────────────────────
	roleSvc := roles.NewRoleService(db)
	roleCtrl := roles.NewRoleController(roleSvc)
	roleGroup := protected.Group("/roles")
	{
		roleGroup.GET("/", roleCtrl.GetAllRoles)
		roleGroup.GET("/:id", roleCtrl.FindOne)
		roleGroup.POST("/", roleCtrl.CreateRole)
		roleGroup.PATCH("/:id", roleCtrl.UpdateRole)
		roleGroup.PUT("/", roleCtrl.SaveRole)
		roleGroup.DELETE("/:id", roleCtrl.DeleteRole)
	}
}

// SetupTenantRoutes kept for backward compatibility.
func SetupTenantRoutes(r *gin.Engine) {
	SetupRoutes(r, database.DB, database.Cache)
}
