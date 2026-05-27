package regions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)

type RegionController struct {
	regionService *RegionService
}

func NewRegionController(regionService *RegionService) *RegionController {
	return &RegionController{regionService: regionService}
}

/* CREATE */

func (rc *RegionController) CreateRegion(c *gin.Context) {
	var createRegionDto dto.CreateRegionDto
	if !middleware.BindAndValidate(c, &createRegionDto) {
		return
	}
	region, err := rc.regionService.Create(c.Request.Context(), &createRegionDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"region": region})
}

func (rc *RegionController) InsertRegions(c *gin.Context) {
	var createRegionDtos []dto.CreateRegionDto
	if err := c.ShouldBindJSON(&createRegionDtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}
	regions, err := rc.regionService.InsertRegions(c.Request.Context(), &createRegionDtos)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"regions": regions})
}

/* UPDATE */

func (rc *RegionController) UpdateRegion(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	var updateRegionDto dto.UpdateRegionDto
	if !middleware.BindAndValidate(c, &updateRegionDto) {
		return
	}
	region, err := rc.regionService.Update(c.Request.Context(), uint(regionId), &updateRegionDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": region})
}

func (rc *RegionController) SaveRegion(c *gin.Context) {
	var region models.Region
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}
	saved, err := rc.regionService.Save(c.Request.Context(), &region)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": saved})
}

/* GET */

func (rc *RegionController) GetAllRegions(c *gin.Context) {
	if c.Query("findOptions") != "" {
		findOptions := make(map[string]any)
		if err := c.ShouldBindQuery(&findOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid find options"})
			return
		}
		regions, total, err := rc.regionService.FindAllWithOptions(c.Request.Context(), findOptions)
		if err != nil {
			global.RespondWithError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"regions": regions, "total": total})
		return
	}
	regions, err := rc.regionService.GetAllRegions(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

func (rc *RegionController) FindOne(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	region, err := rc.regionService.FindOne(c.Request.Context(), uint(regionId))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": region})
}

func (rc *RegionController) FindByName(c *gin.Context) {
	region, err := rc.regionService.FindByRegionName(c.Request.Context(), c.Param("name"))
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"region": region})
}

func (rc *RegionController) GetTenantAssignableRegionsInfo(c *gin.Context) {
	regions, err := rc.regionService.GetTenantAssignableRegionsInfo(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

/* DELETE */

func (rc *RegionController) DeleteRegion(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	if err := rc.regionService.Delete(c.Request.Context(), uint(regionId)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Region deleted successfully"})
}

/* ASSOCIATIONS */

func (rc *RegionController) AddTenantConfigDetailById(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	tcdId, err := strconv.ParseUint(c.Param("tenantConfigDetailId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant config detail ID"})
		return
	}
	if err := rc.regionService.AddTenantConfigDetailById(c.Request.Context(), uint(regionId), uint(tcdId)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail added to region successfully"})
}

func (rc *RegionController) AddTenantConfigDetailsById(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	ids, err := parseUintList(c.Query("tenantConfigDetailIds"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	tcds, err := rc.regionService.AddTenantConfigDetailsById(c.Request.Context(), uint(regionId), ids)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config details added", "tenantConfigDetails": tcds})
}

func (rc *RegionController) RemoveTenantConfigDetailById(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	tcdId, err := strconv.ParseUint(c.Param("tenantConfigDetailId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid tenant config detail ID"})
		return
	}
	if err := rc.regionService.RemoveTenantConfigDetailById(c.Request.Context(), uint(regionId), uint(tcdId)); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail removed from region successfully"})
}

func (rc *RegionController) RemoveTenantConfigDetailsById(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Param("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid region ID"})
		return
	}
	ids, err := parseUintList(c.Query("tenantConfigDetailIds"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if err := rc.regionService.RemoveTenantConfigDetailsById(c.Request.Context(), uint(regionId), ids); err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tenant config details removed from region successfully"})
}

func parseUintList(raw string) ([]uint, error) {
	if raw == "" {
		return nil, fmt.Errorf("missing tenant config detail IDs")
	}
	parts := strings.Split(raw, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", p)
		}
		ids = append(ids, uint(v))
	}
	return ids, nil
}
