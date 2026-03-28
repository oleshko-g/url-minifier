// Package pool provides generic [Pool] of structs that can Reset() themselves before returning into the pool
package pool

import "sync"

// New returns a new [Pool] of structs that can Reset() themselves before returning into the pool
func New[T resetter]() *Pool[T] {
	return &Pool[T]{
		internal: sync.Pool{
			New: NewResetterToAny[T],
		},
	}
}

// resetter is an interface that allows a struct to be reset before returning into the pool
type resetter interface {
	Reset()
}

// NewResetterToAny returns a new value of type T, wrapped as an any
func NewResetterToAny[T resetter]() any {
	var t T
	return any(t)
}

// Pool is a generic struct that can reset itself
type Pool[T resetter] struct {
	internal sync.Pool
}

// Get returns a value from the pool, resetting it first if necessary
func (p *Pool[T]) Get() T {
	return p.internal.Get().(T)
}

// Put puts a value into the pool, resetting it first
func (p *Pool[T]) Put(t T) {
	t.Reset()
	p.internal.Put(t)
}
