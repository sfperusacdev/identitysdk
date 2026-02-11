package staticstore

import "sync"

type StaticStore[K comparable, V any] struct {
	values map[K]V
	mu     sync.RWMutex
}

func New[K comparable, V any]() *StaticStore[K, V] {
	return &StaticStore[K, V]{
		values: make(map[K]V),
	}
}

func (s *StaticStore[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	v, ok := s.values[key]
	s.mu.RUnlock()
	return v, ok
}

func (s *StaticStore[K, V]) Set(key K, value V) {
	s.mu.Lock()
	s.values[key] = value
	s.mu.Unlock()
}

func (s *StaticStore[K, V]) Delete(key K) {
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
}

func (s *StaticStore[K, V]) Len() int {
	s.mu.RLock()
	n := len(s.values)
	s.mu.RUnlock()
	return n
}
