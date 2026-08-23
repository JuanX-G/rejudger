// Custom sync provides a small synchronization helper.
// Written mainly to facilitate shorter locks on global maps.
// You can obtain a point and release a lock on a map while still
// having the safety of a mutex over the value.
package customSync

import (
	"sync"
)

// A lock over a value of type T
type LockedValue[T any] struct {
	v  T
	mu sync.RWMutex
}

// Locks the value readonly and calls f on a point to it.
// The caller must ensure that f does not muatate the value.
func (p *LockedValue[T]) WithReadPointer(f func(*T)) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	f(&p.v)
}

// Locks the value with a read-write pointer
// and calls f on a pointer to the value.
func (p *LockedValue[T]) WithWritePointer(f func(*T)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(&p.v)
}

func MakeLockedValue[T any](v T) *LockedValue[T] {
	return &LockedValue[T]{v: v, mu: sync.RWMutex{}}
}
