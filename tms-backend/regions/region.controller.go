package regions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/gin-gonic/gin"
)


type RegionController struct {
	regionService *RegionService
}

func NewRegionController(regionService *RegionService) *RegionController {
	return &RegionController{
		regionService: regionService,
	}
}


/* Create */

// CreateRegion handles POST request for creating a new region
func (rc *RegionController) CreateRegion(c *gin.Context) {
	var createRegionDto dto.CreateRegionDto
	if err := c.ShouldBindJSON(&createRegionDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	region, err := rc.regionService.Create(&createRegionDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"region": region})
}

// InsertRegions handles POST request for bulk inserting regions
func (rc *RegionController) InsertRegions(c *gin.Context) {
	var createRegionDtos []dto.CreateRegionDto
	if err := c.ShouldBindJSON(&createRegionDtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	regions, err := rc.regionService.InsertRegions(&createRegionDtos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"regions": regions})
}

/* Update */

// UpdateRegion handles PATCH request for updating a region
func (rc *RegionController) UpdateRegion(c *gin.Context) {
	// Convert id to uint
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	var updateRegionDto dto.UpdateRegionDto
	if err := c.ShouldBindJSON(&updateRegionDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	region, err := rc.regionService.Update(uint(regionId), &updateRegionDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"region": region})
}

// SaveRegion handles PUT request for full region update
func (rc *RegionController) SaveRegion(c *gin.Context) {
	var region models.Region
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	savedRegion, err := rc.regionService.Save(&region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"region": savedRegion})
}


/* GET */

// GetAllRegions handles GET request for fetching all regions
func (rc *RegionController) GetAllRegions(c *gin.Context) {

	// Check if query parameters exist
	findOptionsParam := c.Query("findOptions")
	if findOptionsParam != "" {
		// Parse the find options
		findOptions := make(map[string]any)
		err := c.ShouldBindQuery(&findOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid find options"})
			return
		}

		regions, totalCount, err := rc.regionService.FindAllWithOptions(findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"regions": regions,
			"total": totalCount,
		})
		return
	}

	// No query parameters, get all regions
	regions, err := rc.regionService.GetAllRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

// FindOne handles GET request for fetching a single region by ID
func (rc *RegionController) FindOne(c *gin.Context) {
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	region, err := rc.regionService.FindOne(uint(regionId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"region": region})
}

// FindByName handles GET request for fetching a single region by name
func (rc *RegionController) FindByName(c *gin.Context) {
	name := c.Param("name")
	region, err := rc.regionService.FindByRegionName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"region": region})
}

// GetTenantAssignableRegionsInfo handles GET request for fetching regions info for tenant assignment
func (rc *RegionController) GetTenantAssignableRegionsInfo(c *gin.Context) {
	regions, err := rc.regionService.GetTenantAssignableRegionsInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions})
}


/* DELETE */

// DeleteRegion handles DELETE request for deleting a region
func (rc *RegionController) DeleteRegion(c *gin.Context) {
	// Convert id to uint
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	err = rc.regionService.Delete(uint(regionId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Region deleted successfully"})
}


/* ASSOCIATION section */
// AddTenantConfigDetailById handles PATCH request for adding a tenant config detail to a region

func (rc *RegionController) AddTenantConfigDetailById(c *gin.Context) {
	// Parse regionId from URL parameter
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
        return
    }

    // Parse tenantConfigDetailId from URL parameter
    tenantConfigDetailId, err := strconv.ParseUint(c.Params.ByName("tenantConfigDetailId"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant config detail ID"})
        return
    }

	// Call the service method to add the tenant config detail to the region
    err = rc.regionService.AddTenantConfigDetailById(uint(regionId), uint(tenantConfigDetailId))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Tenant config detail added to region successfully",
    })
}

// AddTenantConfigDetailsById handles PATCH request for adding multiple tenant config details to a region
func (rc *RegionController) AddTenantConfigDetailsById(c *gin.Context) {
	// Parse regionId from URL parameter
    regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
        return
    }

	// Get tenant config detail IDs from query string
    tenantConfigDetailIdsStr := c.Query("tenantConfigDetailIds")
    if tenantConfigDetailIdsStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tenant config detail IDs"})
        return
    }

	// Parse tenant config detail IDs from comma-separated string
    tenantConfigDetailIds := []uint{}
    idStrs := strings.Split(tenantConfigDetailIdsStr, ",")
    for _, idStr := range idStrs {
        id, err := strconv.ParseUint(idStr, 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid tenant config detail ID: %s", idStr)})
            return
        }
        tenantConfigDetailIds = append(tenantConfigDetailIds, uint(id))
    }

	// Call the service method to add the tenant config details to the region
    tenantConfigDetails, err := rc.regionService.AddTenantConfigDetailsById(uint(regionId), tenantConfigDetailIds)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Return the updated tenant config details
    c.JSON(http.StatusOK, gin.H{
		"message": "Tenant config details added to region successfully",
		"tenantConfigDetails": tenantConfigDetails,
	})


}



// RemoveTenantConfigDetailById handles DELETE request for removing a tenant config detail from a region
func (rc *RegionController) RemoveTenantConfigDetailById(c *gin.Context) {
	// Get region ID
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	// Get tenant config detail ID
	tenantConfigDetailId, err := strconv.ParseUint(c.Params.ByName("tenantConfigDetailId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant config detail ID"})
		return
	}

	// Remove the tenant config detail from the region
	err = rc.regionService.RemoveTenantConfigDetailById(uint(regionId), uint(tenantConfigDetailId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant config detail removed from region successfully"})
}

// RemoveTenantConfigDetailsById handles DELETE request for removing multiple tenant config details from a region
func (rc *RegionController) RemoveTenantConfigDetailsById(c *gin.Context) {
	// Get region ID
	regionId, err := strconv.ParseUint(c.Params.ByName("regionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	// Get tenant config detail IDs from query string
	tenantConfigDetailIdsStr := c.Query("tenantConfigDetailIds")
	if tenantConfigDetailIdsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tenant config detail IDs"})
		return
	}

	// Parse tenant config detail IDs
	tenantConfigDetailIds := []uint{}
	idStrs := strings.Split(tenantConfigDetailIdsStr, ",")
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid tenant config detail ID: %s", idStr)})
			return
		}
		tenantConfigDetailIds = append(tenantConfigDetailIds, uint(id))
	}

	// Remove the tenant config details from the region
	err = rc.regionService.RemoveTenantConfigDetailsById(uint(regionId), tenantConfigDetailIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant config details removed from region successfully"})
}
