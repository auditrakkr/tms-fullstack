package main

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/regions"
	"github.com/auditrakkr/tms-fullstack/tms-backend/roles"
	tenantconfigdetails "github.com/auditrakkr/tms-fullstack/tms-backend/tenant-config-details"
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants"
	"github.com/auditrakkr/tms-fullstack/tms-backend/users"
	"github.com/gin-gonic/gin"
)

func SetupTenantRoutes(router *gin.Engine) {
	tenantController := tenants.NewTenantController(tenants.NewTenantService())

	tenantGroup := router.Group("/tenants")
	{
		tenantGroup.GET("/", tenantController.GetAllTenants)
		tenantGroup.GET("/:id", tenantController.FindOne)
		tenantGroup.GET("/get-active-tenants-in-region/:regionName", tenantController.FindActiveTenantsByRegionName)
		tenantGroup.POST("/", tenantController.CreateTenant)
		tenantGroup.DELETE("/:id", tenantController.DeleteTenant)
		tenantGroup.PATCH("/:id", tenantController.UpdateTenant)
		/* tenantGroup.POST("/", createTenant)
		tenantGroup.PUT("/:id", updateTenant)
		tenantGroup.DELETE("/:id", deleteTenant) */
	}


	// Add other routes as needed
	userController := users.NewUserController(users.NewUserService())

	userGroup := router.Group("/users")
	{
		userGroup.GET("/", userController.GetAllUsers)
		userGroup.GET("/:id", userController.FindOne)
		userGroup.POST("/", userController.CreateUser)
		userGroup.DELETE("/:id", userController.DeleteUser)
		userGroup.PATCH("/:id", userController.UpdateUser)
	}

	regionController := regions.NewRegionController(regions.NewRegionService())
	regionGroup := router.Group("/regions")
	{
		regionGroup.GET("/", regionController.GetAllRegions)
		regionGroup.POST("/", regionController.CreateRegion)
		regionGroup.PATCH("/:id", regionController.UpdateRegion)
		regionGroup.DELETE("/:id", regionController.DeleteRegion)
	}

	tenantenantConfigDetailsController := tenantconfigdetails.NewTenantConfigDetailsController(tenantconfigdetails.NewTenantConfigDetailsService())
	tenantConfigDetailsGroup := router.Group("/tenant-config-details")
	{
		tenantConfigDetailsGroup.GET("/", tenantenantConfigDetailsController.GetAllTenantConfigDetails)
		tenantConfigDetailsGroup.GET("/:id", tenantenantConfigDetailsController.FindOne)
		tenantConfigDetailsGroup.POST("/", tenantenantConfigDetailsController.CreateTenantConfigDetail)
		tenantConfigDetailsGroup.PATCH("/:id", tenantenantConfigDetailsController.Update)
		tenantConfigDetailsGroup.DELETE("/:id", tenantenantConfigDetailsController.Delete)
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


