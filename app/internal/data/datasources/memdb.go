package datasources

import (
	"sync"

	"github.com/luiscib3r/shortly/app/internal/domain/entities"
)

type MemDB[T entities.Entity] struct {
	store map[string]T
	*sync.RWMutex
}

func NewMemDB[T entities.Entity]() *MemDB[T] {
	return &MemDB[T]{
		store:   make(map[string]T),
		RWMutex: &sync.RWMutex{},
	}
}

func (s MemDB[T]) FindAll() []T {
	var all []T
	s.RLock()
	for _, v := range s.store {
		all = append(all, v)
	}
	s.RUnlock()
	return all
}

func (s MemDB[T]) FindById(id string) (T, bool) {
	s.RLock()
	v, ok := s.store[id]
	s.RUnlock()
	return v, ok
}

func (s *MemDB[T]) Save(entity T) T {
	s.Lock()
	s.store[entity.Id()] = entity
	result := s.store[entity.Id()]
	s.Unlock()
	return result
}

func (s *MemDB[T]) Delete(id string) bool {
	s.Lock()
	delete(s.store, id)
	s.Unlock()
	return true
}

func (s MemDB[T]) Count() int {
	return len(s.store)
}
