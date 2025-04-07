package regions

import (
	"strconv"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
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

func (rc *RegionController) CreateRegion(c *gin.Context) {
	var createRegionDto dto.CreateRegionDto
	if err := c.ShouldBindJSON(&createRegionDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	region, err := rc.regionService.Create(&createRegionDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"region": region})
}

/* Update */

func (rc *RegionController) UpdateRegion(c *gin.Context) {
	// Convert id to uint
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid region ID"})
		return
	}

	var updateRegionDto dto.UpdateRegionDto
	if err := c.ShouldBindJSON(&updateRegionDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	region, err := rc.regionService.Update(uint(id), &updateRegionDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"region": region})
}


/* GET */
func (rc *RegionController) GetAllRegions(c *gin.Context) {
	regions, err := rc.regionService.GetAllRegions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"regions": regions})
}

/* DELETE */
func (rc *RegionController) DeleteRegion(c *gin.Context) {
	// Convert id to uint
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid region ID"})
		return
	}
	err = rc.regionService.Delete(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Region deleted successfully"})
}