package customSync

import (
	"sync"
)

type PointerLock[T any] struct {
	p T
	mu sync.RWMutex
}

func (p *PointerLock[T]) WithReadPointer(f func(*T)) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	f(&p.p)
}

func (p *PointerLock[T]) WithWritePointer(f func(*T)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(&p.p)
}

func MakePointerLock[T any](v T) *PointerLock[T] {
	return &PointerLock[T]{p: v, mu: sync.RWMutex{}}
}
