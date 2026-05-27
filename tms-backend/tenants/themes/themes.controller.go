package themes

import (
	"net/http"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/middleware"
	"github.com/gin-gonic/gin"
)

type ThemeController struct {
	themeService *ThemeService
}

func NewThemeController(themeService *ThemeService) *ThemeController {
	return &ThemeController{themeService: themeService}
}

func (tc *ThemeController) CreateTheme(c *gin.Context) {
	var createThemeDto dto.CreateThemeDto
	if !middleware.BindAndValidate(c, &createThemeDto) {
		return
	}
	theme, err := tc.themeService.CreateTheme(c.Request.Context(), &createThemeDto)
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"theme": theme})
}

func (tc *ThemeController) FindAll(c *gin.Context) {
	themes, err := tc.themeService.FindAll(c.Request.Context())
	if err != nil {
		global.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"themes": themes})
}
