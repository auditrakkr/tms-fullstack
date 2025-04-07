package themes

import (
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/gin-gonic/gin"
)


type ThemeController struct {
	themeService *ThemeService
}

func NewThemeController(themeService *ThemeService) *ThemeController {
	return &ThemeController{
		themeService: themeService,
	}
}


/* CREATE */
func (tc *ThemeController) CreateTheme(c *gin.Context) {
	var createThemeDto dto.CreateThemeDto
	if err := c.ShouldBindJSON(&createThemeDto); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	theme, err := tc.themeService.CreateTheme(&createThemeDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"theme": theme})
}


func (tc *ThemeController) FindAll(c *gin.Context) {
	c.String(200, "This is themes")
}