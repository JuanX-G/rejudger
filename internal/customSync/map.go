package customSync

import "sync"

type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	base map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{mu: sync.RWMutex{}, base: make(map[K]V)}
}

func (sm *SafeMap[K, V]) WithRlock(k K, fun func(V)) bool {
	sm.mu.RLock()
	defer sm.mu.Unlock()
	if v, ok := sm.base[k]; !ok {
		return false
	} else {
		fun(v)
		return true
	}
}

func (sm *SafeMap[K, V]) WithLock(k K, fun func(V)) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if v, ok := sm.base[k]; !ok {
		return false
	} else {
		fun(v)
		return true
	}
}

func (sm *SafeMap[K, V]) Store(k K, v V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.base[k] = v
}

func (sm *SafeMap[K, V]) Delete(k K) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, ok := sm.base[k]; !ok {
		return false
	} else {
		delete(sm.base, k)
		return true
	}
}

func (sm *SafeMap[K, V]) Len() int {
	return len(sm.base)
}

func (sm *SafeMap[K, V]) For(fun func(K, V)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for k, v := range sm.base {
		fun(k, v)
	}
}
