package themes

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
)


type ThemeService struct {
	themeRepo repositories.Repository[models.Theme]
}


func NewThemeService() *ThemeService {
	return &ThemeService{
		themeRepo: repositories.Repository[models.Theme]{DB: database.DB},
	}
}


func (s *ThemeService) CreateTheme(createThemeDto *dto.CreateThemeDto) (*models.Theme, error) {
	newTheme := &models.Theme{}
	if err := copier.Copy(newTheme, createThemeDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %v", err)
	}

	newTheme, err := s.themeRepo.Create(newTheme)
	if err != nil {
		return nil, fmt.Errorf("failed to create theme: %v", err)
	}
	return newTheme, nil
}