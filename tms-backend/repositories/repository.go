package repositories

import (
	"errors"

	"gorm.io/gorm"
)

// GormRepository is the concrete GORM-backed implementation of Repository[T].
// Services declare their repo fields using the Repository[T] interface so they
// can be tested with fakes; constructors wire in *GormRepository[T].
type GormRepository[T any] struct {
	DB *gorm.DB
}

// NewRepository constructs a GormRepository for the given type T.
func NewRepository[T any](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{DB: db}
}

func (r *GormRepository[T]) Create(entity *T) (*T, error) {
	if err := r.DB.Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormRepository[T]) FindByID(id uint) (*T, error) {
	var entity T
	if err := r.DB.First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindAll() ([]T, error) {
	var entities []T
	if err := r.DB.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GormRepository[T]) FindOne(condition map[string]any) (*T, error) {
	var entity T
	if err := r.DB.Where(condition).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) Update(entity *T) error {
	return r.DB.Model(entity).Updates(entity).Error
}

func (r *GormRepository[T]) Delete(id uint) error {
	result := r.DB.Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository[T]) Save(entity *T) (*T, error) {
	if err := r.DB.Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormRepository[T]) Find(condition map[string]any) ([]T, error) {
	var entities []T
	if err := r.DB.Where(condition).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// FindAndCount runs count and find as separate queries on the same conditions,
// avoiding the GORM v2 bug where chaining Count after Find is a no-op.
func (r *GormRepository[T]) FindAndCount(condition map[string]any) ([]T, int64, error) {
	var entities []T
	var count int64

	base := r.DB.Model(new(T)).Where(condition)
	if err := base.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := base.Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	return entities, count, nil
}

func (r *GormRepository[T]) Remove(entity *T) (T, error) {
	if err := r.DB.Delete(entity).Error; err != nil {
		return *new(T), err
	}
	return *entity, nil
}

func (r *GormRepository[T]) CreateQueryBuilder() *gorm.DB {
	return r.DB.Model(new(T))
}
