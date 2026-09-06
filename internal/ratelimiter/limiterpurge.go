package ratelimiter

import (
	"slices"
	"time"
)

func (l *Limiter[T]) purge() {
	for {
		select {
		case <-l.done:
			l.cleanUpTicker.Stop()
			l.stopped.Store(true)
			return
		case <-l.cleanUpTicker.C:
			l.runCleanup()
		}
	}
}

func (l *Limiter[T]) runCleanup() {
	now := time.Now()

	var expiredKeys []T
	l.store.For(func(k T, data *limitData) {
		if data.evict || now.After(data.lastRecorded.Add(l.timeToEvict)) {
			expiredKeys = append(expiredKeys, k)
		}
	})

	for _, k := range expiredKeys {
		l.store.Delete(k)
	}

	currentSize := l.store.Len()
	if currentSize <= l.maxSize {
		return
	}

	type entry struct {
		key          T
		lastRecorded time.Time
	}

	entries := make([]entry, 0, currentSize)
	l.store.For(func(k T, data *limitData) {
		entries = append(entries, entry{
			key:          k,
			lastRecorded: data.lastRecorded,
		})
	})

	slices.SortFunc(entries, func(a, b entry) int {
		return a.lastRecorded.Compare(b.lastRecorded)
	})

	excess := len(entries) - l.maxSize
	if excess <= 0 {
		return
	}

	for i := range excess {
		l.store.Delete(entries[i].key)
	}
}
