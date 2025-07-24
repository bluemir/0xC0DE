package datastruct

import (
	"sync"
)

type null struct{}

type Set[T comparable] struct {
	internal *sync.Map
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		internal: &sync.Map{},
	}
}

// Add add
func (s *Set[T]) Add(item T) {

	s.internal.Store(item, null{})
}

// Remove deletes the specified item from the map
func (s *Set[T]) Remove(item T) {
	s.internal.Delete(item)
}

// Has looks for the existence of an item
func (s *Set[T]) Has(item T) bool {
	_, ok := s.internal.Load(item)
	return ok
}

// Clear removes all items from the set
func (s *Set[T]) Clear() {
	s.internal.Clear()
}

// ForEach run function each item, with lock. return error when occur error
func (s *Set[T]) ForEach(fn func(T) error) error {
	var err error
	s.internal.Range(func(k, v any) bool {
		err = fn(k.(T))
		if err != nil {
			return false
		}
		return true
	})
	return err
}
