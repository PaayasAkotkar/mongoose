// Package stack Simple Stack implementation
package stack

import (
	"slices"
	"sync"
)

// Stack -> new entry at first, old entry at last
type Stack[T comparable] struct {
	mu   sync.Mutex
	data []T
}

func NewStack[T comparable]() *Stack[T] {
	return &Stack[T]{}
}

// get funcs
func (s *Stack[T]) len() int    { return len(s.data) }
func (s *Stack[T]) empty() bool { return len(s.data) == 0 }

func (s *Stack[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.len()
}

func (s *Stack[T]) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.empty()
}

// Latest returns a copy of the newest entry
func (s *Stack[T]) Latest() *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() {
		return nil
	}
	v := s.data[0]
	return &v
}

// Oldest returns a copy of the oldest entry
func (s *Stack[T]) Oldest() *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() {
		return nil
	}
	v := s.data[len(s.data)-1]
	return &v
}

func (s *Stack[T]) Max() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]T, s.len())
	copy(out, s.data)
	return out
}

// Range returns a copy of up to limit entries
func (s *Stack[T]) Range(limit int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit < 0 {
		limit = 0
	}
	if limit > s.len() {
		limit = s.len()
	}
	out := make([]T, limit)
	copy(out, s.data[:limit])
	return out
}

// end

// set funcs

// Push inserts at the front (newest first)
func (s *Stack[T]) Push(name T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append([]T{name}, s.data...)
}

func (s *Stack[T]) Erase(name T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.data {

		if r == name {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return
		}
	}
}

// Rearrange moves the [from, to] range to the front
// swap the position from this rank to this rank
// example: [5,4,3,2,1] -> [3,4] -> [2,1,5,4,3]
// note: toIndex is considered whole number -> if 4 than 4+1 than 4 be careful
func (s *Stack[T]) Rearrange(from int, to int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if from > s.len() || to > s.len() {
		return
	}

	priority := make([]T, to-from+1)
	copy(priority, s.data[from:to+1])

	remaining := make([]T, 0, to-from+1)
	remaining = append(remaining, s.data[:from]...)
	remaining = append(remaining, s.data[to+1:]...)

	s.data = append(priority, remaining...)
}

// PushLast current rank-0 to last
// example: [5,4,3,2,1] -> [4,3,2,1,5]
func (s *Stack[T]) PushLast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() {
		return
	}

	first := s.data[0]
	copy(s.data, s.data[1:])
	s.data[s.len()-1] = first
}

func (s *Stack[T]) Replace(index int, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index > s.len() {
		return
	}
	s.data[index] = value
}

func (s *Stack[T]) Contains(data T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Contains(s.data, data)
}

// end
