package themes

import (
	"context"
	"fmt"

	dto "github.com/auditrakkr/tms-fullstack/tms-backend/dtos"
	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/auditrakkr/tms-fullstack/tms-backend/repositories"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type ThemeService struct {
	themeRepo repositories.Repository[models.Theme]
}

func NewThemeService(db *gorm.DB) *ThemeService {
	return &ThemeService{
		themeRepo: repositories.NewRepository[models.Theme](db),
	}
}

func (s *ThemeService) CreateTheme(ctx context.Context, createThemeDto *dto.CreateThemeDto) (*models.Theme, error) {
	newTheme := &models.Theme{}
	if err := copier.Copy(newTheme, createThemeDto); err != nil {
		return nil, fmt.Errorf("failed to map dto: %w", err)
	}
	newTheme, err := s.themeRepo.Create(newTheme)
	if err != nil {
		return nil, fmt.Errorf("failed to create theme: %w", err)
	}
	return newTheme, nil
}

func (s *ThemeService) FindAll(ctx context.Context) ([]models.Theme, error) {
	themes, err := s.themeRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get themes: %w", err)
	}
	return themes, nil
}
