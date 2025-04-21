package main

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/regions"
	"github.com/auditrakkr/tms-fullstack/tms-backend/roles"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenant-config-details"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants/billings"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants/themes"
	"github.com/auditrakkr/tms-fullstack/tms-backend/users"
	"github.com/gin-gonic/gin"
)

func SetupTenantRoutes(router *gin.Engine) {
	tenantController := tenants.NewTenantController(tenants.NewTenantService())
	themeController := themes.NewThemeController(themes.NewThemeService())
	billingController := billings.NewBillingController(billings.NewBillingService())

	tenantGroup := router.Group("/tenants")
	{
		tenantGroup.GET("/", tenantController.GetAllTenants)
		tenantGroup.GET("/:id", tenantController.FindOne)
		tenantGroup.GET("/get-active-tenants-in-region/:regionName", tenantController.FindActiveTenantsByRegionName)
		tenantGroup.GET("/themes", themeController.FindAll)
		tenantGroup.GET("/billings", billingController.FindAll)

		tenantGroup.POST("/", tenantController.CreateTenant)
		tenantGroup.POST("/themes", themeController.CreateTheme)
		tenantGroup.POST("/billings", billingController.CreateBilling)


		tenantGroup.PATCH("/:id", tenantController.UpdateTenant)
		tenantGroup.DELETE("/:id", tenantController.DeleteTenant)
	}


	// Add other routes as needed
	userController := users.NewUserController(users.NewUserService())

	userGroup := router.Group("/users")
	{
		userGroup.GET("/", userController.GetAllUsers)
		userGroup.GET("/:id", userController.FindOne)

		userGroup.POST("/", userController.CreateUser)

		userGroup.PATCH("/:id", userController.UpdateUser)
		userGroup.DELETE("/:id", userController.DeleteUser)
	}

	regionController := regions.NewRegionController(regions.NewRegionService())
	regionGroup := router.Group("/regions")
	{
		regionGroup.GET("/", regionController.GetAllRegions)
		regionGroup.GET("/:regionId", regionController.FindOne)
		regionGroup.GET("/by-name/:name", regionController.FindByName)
		regionGroup.GET("/get-tenant-assignable-regions-info", regionController.GetTenantAssignableRegionsInfo)

		regionGroup.POST("/", regionController.CreateRegion)
		regionGroup.POST("/insert", regionController.InsertRegions)

		regionGroup.PATCH("/:regionId", regionController.UpdateRegion)
		regionGroup.PUT("/", regionController.SaveRegion)
		regionGroup.DELETE("/:regionId", regionController.DeleteRegion)

		// Association endpoints
		regionGroup.PATCH("/:regionId/tenant-config-detail/:tenantConfigDetailId", regionController.AddTenantConfigDetailById)
		regionGroup.PATCH("/:regionId/tenant-config-details", regionController.AddTenantConfigDetailsById)
		regionGroup.DELETE("/:regionId/tenant-config-detail/:tenantConfigDetailId", regionController.RemoveTenantConfigDetailById)
		regionGroup.DELETE("/:regionId/tenant-config-details", regionController.RemoveTenantConfigDetailsById)
	}

	tenantConfigDetailsController := tenantconfigdetails.NewTenantConfigDetailsController(tenantconfigdetails.NewTenantConfigDetailsService())
	tenantConfigDetailsGroup := router.Group("/tenant-config-details")
	{
		tenantConfigDetailsGroup.GET("/", tenantConfigDetailsController.GetAllTenantConfigDetails)
		tenantConfigDetailsGroup.GET("/:id", tenantConfigDetailsController.FindOne)

		tenantConfigDetailsGroup.POST("/", tenantConfigDetailsController.CreateTenantConfigDetail)

		tenantConfigDetailsGroup.PATCH("/:id", tenantConfigDetailsController.Update)
		tenantConfigDetailsGroup.DELETE("/:id", tenantConfigDetailsController.Delete)
	}

	roleController := roles.NewRoleController(roles.NewRoleService())
	roleGroup := router.Group("/roles")
	{
		roleGroup.GET("/", roleController.GetAllRoles)
		roleGroup.GET("/:id", roleController.FindOne)

		roleGroup.POST("/", roleController.CreateRole)

		roleGroup.PATCH("/:id", roleController.UpdateRole)
		roleGroup.PUT("/", roleController.SaveRole)
		roleGroup.DELETE("/:id", roleController.DeleteRole)
	}

}


