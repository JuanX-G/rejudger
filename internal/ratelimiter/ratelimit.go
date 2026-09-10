package ratelimiter

import (
	"revit/internal/customSync"
	"sync/atomic"
	"time"
)

type RateLimiterConfig struct {
	TimeToEvict, CleanupTime time.Duration
	MaxSize, MaxCounter      int
}

type limitData struct {
	lastRecorded time.Time
	counter      int
	evict        bool
}

type Limiter[T comparable] struct {
	store         *customSync.SafeMap[T, *limitData]
	timeToEvict   time.Duration
	maxSize       int
	maxCounter    int
	done          chan struct{}
	stopped       atomic.Bool
	cleanUpTicker *time.Ticker
}

func NewLimiter[T comparable](cfg RateLimiterConfig) *Limiter[T] {
	lim := &Limiter[T]{
		store:         customSync.NewSafeMap[T, *limitData](),
		timeToEvict:   cfg.TimeToEvict,
		maxSize:       cfg.MaxSize,
		maxCounter:    cfg.MaxCounter,
		done:          make(chan struct{}),
		cleanUpTicker: time.NewTicker(cfg.CleanupTime),
	}

	go lim.purge()
	return lim
}

func (l *Limiter[T]) Allow(k T) bool {
	now := time.Now()
	allowed := false

	found := l.store.WithLock(k, func(data *limitData) {
		if now.After(data.lastRecorded.Add(l.timeToEvict)) {
			data.counter = 0
			data.evict = false
		}

		if data.counter < l.maxCounter {
			data.counter++
			data.lastRecorded = now
			allowed = true
		}
	})

	if !found {
		l.store.Store(k, &limitData{
			lastRecorded: now,
			counter:      1,
		})
		return true
	}

	return allowed
}

func (l *Limiter[T]) Stop() {
	close(l.done)
}
