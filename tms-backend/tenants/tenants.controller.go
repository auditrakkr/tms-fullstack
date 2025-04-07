package tenants

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/gin-gonic/gin"
)

type TenantController struct {
	tenantService *TenantService
}
func NewTenantController(tenantService *TenantService) *TenantController {
	return &TenantController{
		tenantService: tenantService,
	}
}

/* func CreateTenant(c *gin.Context, createTenantDto dto.CreateTenantDto) {
	// Parse the request body into the CreateTenantDto


	// Call the service to create the tenant
	err := tenantService.CreateTenant(&createTenantDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tenant created successfully"})
}
 */

 func (tc *TenantController) CreateTenant(c *gin.Context) {
	var req Request
	var createTenantDto dto.CreateTenantDto
	if err := c.ShouldBindBodyWithJSON(&createTenantDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
	}

	createPrimaryContact, err := strconv.ParseUint(c.Query("createPrimaryContact"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid createPrimaryContact value"})
	}

	tenant, err := tc.tenantService.CreateTenant(&createTenantDto, uint(createPrimaryContact), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"tenant": tenant})
}

func (tc *TenantController) GetAllTenants(c *gin.Context) {
	tenants, err := tc.tenantService.GetAllTenants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

func (tc *TenantController) FindOne(c *gin.Context) {
	// Convert id to uint
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}
	tenant, err := tc.tenantService.FindOne(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tenant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})

}

func (tc *TenantController) FindActiveTenantsByRegionName(c *gin.Context) {
	regionName := c.Param("regionName")
	tenants, err := tc.tenantService.FindActiveTenantsByRegionName(regionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

/* func updateTenant(c *gin.Context) {
	// Logic to update a tenant by ID
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Tenant updated", "id": id})
}


func deleteTenant(c *gin.Context) {
	// Logic to delete a tenant by ID
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted", "id": id})
} */


func (tc *TenantController) DeleteTenant(c *gin.Context) {
	// Convert id to uint
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}
	err = tc.tenantService.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
	//c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted", "id": id})
}


/* UPDATE */
func (tc *TenantController) UpdateTenant(c *gin.Context) {
	// Convert id to uint
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var updateTenantDto dto.UpdateTenantDto
	if err := c.ShouldBindJSON(&updateTenantDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	tenant, err := tc.tenantService.Update(uint(id), &updateTenantDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenant": tenant})
	c.JSON(http.StatusOK, gin.H{"message": "Tenant updated successfully"})
}