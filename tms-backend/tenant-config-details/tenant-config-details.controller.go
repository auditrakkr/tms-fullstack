package tenantconfigdetails

import (
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/gin-gonic/gin"
)


type TenantConfigDetailsController struct{
	tenantConfigDetailsService *TenantConfigDetailsService
}

func NewTenantConfigDetailsController(tenantConfigDetailsService *TenantConfigDetailsService) *TenantConfigDetailsController {
	return &TenantConfigDetailsController{
		tenantConfigDetailsService: tenantConfigDetailsService,
	}
}


/* CREATE */
func (tc *TenantConfigDetailsController) CreateTenantConfigDetail(c *gin.Context) {
	var createTenantConfigDetailDto dto.CreateTenantConfigDetailDto
	if err := c.ShouldBindJSON(&createTenantConfigDetailDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	tenantConfigDetail, err := tc.tenantConfigDetailsService.CreateTenantConfigDetail(&createTenantConfigDetailDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"tenantConfigDetail": tenantConfigDetail})
}

/* UPDATE */
func (tc *TenantConfigDetailsController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid tenant config detail ID"})
		return
	}
	var updateTenantConfigDetailDto dto.CreateTenantConfigDetailDto
	if err := c.ShouldBindJSON(&updateTenantConfigDetailDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	tenantConfigDetail, err := tc.tenantConfigDetailsService.Update(uint(id), &updateTenantConfigDetailDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"tenantConfigDetail": tenantConfigDetail})

}

/* FIND */
func (tc *TenantConfigDetailsController) GetAllTenantConfigDetails (c *gin.Context) {
	tenantConfigDetails, err := tc.tenantConfigDetailsService.GetAllTenantConfigDetails()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"tenantConfigDetails": tenantConfigDetails})
}

func (tc *TenantConfigDetailsController) FindOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid tenant config detail ID"})
		return
	}
	tenantConfigDetail, err := tc.tenantConfigDetailsService.FindOne(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"tenantConfigDetail": tenantConfigDetail})
}

/* DELETE */
func (tc *TenantConfigDetailsController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid tenant config detail ID"})
		return
	}
	err = tc.tenantConfigDetailsService.Delete(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Tenant config detail deleted successfully"})
}