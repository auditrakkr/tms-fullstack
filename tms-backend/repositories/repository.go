package repositories

import (
	"errors"

	"gorm.io/gorm"
)

// Generic Repository
type Repository[T any] struct {
	DB *gorm.DB // Assuming DB is a gorm.DB instance
	// Add any other fields you need for your repository
}

// Constructor function for Repository
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{DB: db}
}

// Create a new record
func (r *Repository[T]) Create(entity *T) error {
	return r.DB.Create(&entity).Error
}

// Find a record by ID
func (r *Repository[T]) FindByID(id uint) (*T, error) {
	var entity T
	err := r.DB.First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err // Return error to signal "not found"
		}
		return nil, err
	}
	return &entity, nil
}

// Get all records
func (r *Repository[T]) FindAll() ([]T, error) {
	var entities []T
	err := r.DB.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// FindOne - Find a single record based on conditions
func (r *Repository[T]) FindOne(condition map[string]any) (*T, error) {
	var entity T
	err := r.DB.Where(condition).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &entity, nil
}

// Update a record
func (r *Repository[T]) Update(entity *T) error {
	return r.DB.Model(&entity).Updates(entity).Error
}

func (r *Repository[T]) Delete(id uint) error {
	result := r.DB.Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateQueryBuilder - Returns a query instance to build custom queries
func (r *Repository[T]) CreateQueryBuilder() *gorm.DB {
	var entity T
	return r.DB.Model(&entity)
}

// Save (Create or Update)
func (r *Repository[T]) Save(entity *T) (*T, error) {
	err := r.DB.Save(entity).Error
	return entity, err
}

// Find - Find multiple records based on conditions
func (r *Repository[T]) Find(condition map[string]any) ([]T, error) {
	var entities []T
	err := r.DB.Where(condition).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// FindAndCount - Find records and return the count
func (r *Repository[T]) FindAndCount(condition map[string]any) ([]T, int64, error) {
	var entities []T
	var count int64

	query := r.DB.Where(condition)
	err := query.Find(&entities).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return entities, count, nil
}

// Remove - Deletes an entity
func (r *Repository[T]) Remove(entity *T) (T, error) {
	err := r.DB.Delete(entity).Error
	if err != nil {
		return *new(T), err
	}
	return *entity, nil
}
