package store

import "sync"

type Store[T any] struct {
	mu        sync.RWMutex
	value     T
	listeners map[int]func(T)
	nextID    int
}

func NewStore[T any](initial T) *Store[T] {
	return &Store[T]{
		value:     initial,
		listeners: make(map[int]func(T)),
		nextID:    0,
	}
}

func (s *Store[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *Store[T]) Set(value T) {
	s.mu.Lock()
	s.value = value
	listeners := make([]func(T), 0, len(s.listeners))
	for _, l := range s.listeners {
		listeners = append(listeners, l)
	}
	s.mu.Unlock()

	for _, l := range listeners {
		if l != nil {
			l(value)
		}
	}
}

func (s *Store[T]) Update(updater func(T) T) T {
	s.mu.Lock()
	newValue := updater(s.value)
	s.value = newValue
	listeners := make([]func(T), 0, len(s.listeners))
	for _, l := range s.listeners {
		listeners = append(listeners, l)
	}
	s.mu.Unlock()

	for _, l := range listeners {
		if l != nil {
			l(newValue)
		}
	}

	return newValue
}

func (s *Store[T]) Subscribe(listener func(T)) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.listeners[id] = listener
	current := s.value
	s.mu.Unlock()

	if listener != nil {
		listener(current)
	}

	return func() {
		s.mu.Lock()
		delete(s.listeners, id)
		s.mu.Unlock()
	}
}
