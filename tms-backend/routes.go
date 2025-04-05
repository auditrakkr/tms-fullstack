package main

import (
	"github.com/auditrakkr/tms-fullstack/tms-backend/tenants"
	"github.com/gin-gonic/gin"
)

func SetupTenantRoutes(router *gin.Engine) {
	tenantController := tenants.NewTenantController(tenants.NewTenantService())

	tenantGroup := router.Group("/tenants")
	{
		tenantGroup.GET("/", tenantController.GetAllTenants)
		tenantGroup.GET("/:id", tenantController.FindOne)
		tenantGroup.GET("/region/:regionName", tenantController.FindActiveTenantsByRegionName)
		tenantGroup.POST("/", tenantController.CreateTenant)
		tenantGroup.DELETE("/:id", tenantController.DeleteTenant)
		/* tenantGroup.POST("/", createTenant)
		tenantGroup.PUT("/:id", updateTenant)
		tenantGroup.DELETE("/:id", deleteTenant) */
	}
}


