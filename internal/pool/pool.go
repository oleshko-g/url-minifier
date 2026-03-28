// Package pool provides generic [Pool] of structs that can Reset() themselves before returning into the pool
package pool

import 	"sync"

func New[T resetter]() *Pool[T] {
	return &Pool[T]{
		internal: sync.Pool{
			New: func() any {
				var t T
				return any(t)
			},
		},
	}
}

// Pool is a generic struct that can reset itself
type Pool[T resetter] struct {
	internal sync.Pool
}

func (p *Pool[T]) Get() T {
	return p.internal.Get().(T)
}

func (p *Pool[T]) Put(t T) {
	t.Reset()
	p.internal.Put(t)
}

type resetter interface {
	Reset()
}
