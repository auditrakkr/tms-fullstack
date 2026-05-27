package tenantconfigdetails

import (
	"net/http"
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/gin-gonic/gin"
)

type TenantConfigDetailsController struct {
	tenantConfigDetailsService *TenantConfigDetailsService
}

func NewTenantConfigDetailsController(svc *TenantConfigDetailsService) *TenantConfigDetailsController {
	return &TenantConfigDetailsController{tenantConfigDetailsService: svc}
}

/* CREATE */

func (tc *TenantConfigDetailsController) CreateTenantConfigDetail(c *gin.Context) {
	var createDto dto.CreateTenantConfigDetailDto
	if !middleware.BindAndValidate(c, &createDto) {
		return
	}
	detail, err := tc.tenantConfigDetailsService.CreateTenantConfigDetail(c.Request.Context(), &createDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tenantConfigDetail": detail})
}

/* UPDATE */

func (tc *TenantConfigDetailsController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant config detail ID"})
		return
	}
	var updateDto dto.CreateTenantConfigDetailDto
	if !middleware.BindAndValidate(c, &updateDto) {
		return
	}
	detail, err := tc.tenantConfigDetailsService.Update(c.Request.Context(), uint(id), &updateDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenantConfigDetail": detail})
}

/* FIND */

func (tc *TenantConfigDetailsController) GetAllTenantConfigDetails(c *gin.Context) {
	details, err := tc.tenantConfigDetailsService.GetAllTenantConfigDetails(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenantConfigDetails": details})
}

func (tc *TenantConfigDetailsController) FindOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant config detail ID"})
		return
	}
	detail, err := tc.tenantConfigDetailsService.FindOne(c.Request.Context(), uint(id))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tenantConfigDetail": detail})
}

/* DELETE */

func (tc *TenantConfigDetailsController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant config detail ID"})
		return
	}
	if err := tc.tenantConfigDetailsService.Delete(c.Request.Context(), uint(id)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail deleted successfully"})
}
