// Package testutil provides shared helpers for unit and integration tests.
//
// FakeRepository is an in-memory implementation of repositories.Repository[T]
// that lets service-layer tests run without a real database.
// It satisfies the interface so any *Service that accepts Repository[T] can be
// wired with a FakeRepository in tests.
package testutil

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// ErrNotFound is the sentinel returned by FakeRepository when no record matches.
var ErrNotFound = errors.New("record not found")

// FakeRepository is a thread-safe in-memory store that satisfies
// repositories.Repository[T].  IDs are assigned sequentially starting at 1.
type FakeRepository[T any] struct {
	mu      sync.RWMutex
	records map[uint]*T
	nextID  uint
	// IDFunc extracts the uint primary key from a record; defaults to zero.
	IDFunc  func(*T) uint
	// SetIDFunc injects a uint primary key into a record.
	SetIDFunc func(*T, uint)
}

// NewFakeRepository constructs a FakeRepository. Callers typically set
// IDFunc and SetIDFunc to match their model's primary-key field.
func NewFakeRepository[T any]() *FakeRepository[T] {
	return &FakeRepository[T]{
		records: make(map[uint]*T),
		nextID:  1,
	}
}

func (r *FakeRepository[T]) Create(entity *T) (*T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	if r.SetIDFunc != nil {
		r.SetIDFunc(entity, id)
	}
	clone := *entity
	r.records[id] = &clone
	return &clone, nil
}

func (r *FakeRepository[T]) FindByID(id uint) (*T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if v, ok := r.records[id]; ok {
		clone := *v
		return &clone, nil
	}
	return nil, ErrNotFound
}

func (r *FakeRepository[T]) FindAll() ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]T, 0, len(r.records))
	for _, v := range r.records {
		out = append(out, *v)
	}
	return out, nil
}

func (r *FakeRepository[T]) FindOne(condition map[string]any) (*T, error) {
	// Simple linear scan — sufficient for unit tests.
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.records {
		return v, nil // returns first record; override in specific tests if needed
	}
	return nil, ErrNotFound
}

func (r *FakeRepository[T]) Update(entity *T) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.IDFunc == nil {
		return errors.New("FakeRepository: IDFunc not set, cannot Update")
	}
	id := r.IDFunc(entity)
	if _, ok := r.records[id]; !ok {
		return ErrNotFound
	}
	clone := *entity
	r.records[id] = &clone
	return nil
}

func (r *FakeRepository[T]) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.records[id]; !ok {
		return ErrNotFound
	}
	delete(r.records, id)
	return nil
}

func (r *FakeRepository[T]) Save(entity *T) (*T, error) {
	if r.IDFunc == nil {
		return r.Create(entity)
	}
	id := r.IDFunc(entity)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.records[id]; !ok || id == 0 {
		newID := r.nextID
		r.nextID++
		if r.SetIDFunc != nil {
			r.SetIDFunc(entity, newID)
		}
		clone := *entity
		r.records[newID] = &clone
		return &clone, nil
	}
	clone := *entity
	r.records[id] = &clone
	return &clone, nil
}

func (r *FakeRepository[T]) Find(condition map[string]any) ([]T, error) {
	return r.FindAll()
}

func (r *FakeRepository[T]) FindAndCount(condition map[string]any) ([]T, int64, error) {
	all, err := r.FindAll()
	return all, int64(len(all)), err
}

func (r *FakeRepository[T]) Remove(entity *T) (T, error) {
	if r.IDFunc == nil {
		return *new(T), errors.New("FakeRepository: IDFunc not set, cannot Remove")
	}
	id := r.IDFunc(entity)
	if err := r.Delete(id); err != nil {
		return *new(T), err
	}
	return *entity, nil
}

// CreateQueryBuilder returns a nil *gorm.DB; unit tests should not call GORM
// directly. Override this method if you need raw query support in tests.
func (r *FakeRepository[T]) CreateQueryBuilder() *gorm.DB {
	return nil
}
