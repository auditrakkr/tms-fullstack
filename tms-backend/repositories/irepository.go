package repositories

import "gorm.io/gorm"

// Repository is the generic data-access interface consumed by service layers.
// The concrete implementation is GormRepository, defined in repository.go.
// Defining the interface here (near the implementation) allows services to
// accept the interface type while keeping the GORM dependency isolated.
type Repository[T any] interface {
	Create(entity *T) (*T, error)
	FindByID(id uint) (*T, error)
	FindAll() ([]T, error)
	FindOne(condition map[string]any) (*T, error)
	Update(entity *T) error
	Delete(id uint) error
	Save(entity *T) (*T, error)
	Find(condition map[string]any) ([]T, error)
	FindAndCount(condition map[string]any) ([]T, int64, error)
	Remove(entity *T) (T, error)
	// CreateQueryBuilder returns a scoped *gorm.DB for building custom queries
	// (e.g. associations, joins, select projections).
	CreateQueryBuilder() *gorm.DB
}
