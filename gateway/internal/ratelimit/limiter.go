package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type Limiter struct {
	mu sync.Mutex

	capacity   float64
	refillRate float64

	buckets map[string]*bucket

	now func() time.Time
}

func New(
	capacity int,
	refillPerSecond float64,
) *Limiter {
	return &Limiter{
		capacity:   float64(capacity),
		refillRate: refillPerSecond,
		buckets:    make(map[string]*bucket),
		now:        time.Now,
	}
}

func (l *Limiter) Allow(
	key string,
) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{
			tokens:     l.capacity,
			lastRefill: now,
		}

		l.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()

	if elapsed > 0 {
		b.tokens += elapsed * l.refillRate

		if b.tokens > l.capacity {
			b.tokens = l.capacity
		}

		b.lastRefill = now
	}

	if b.tokens >= 1 {
		b.tokens--

		return true, 0
	}

	if l.refillRate <= 0 {
		return false, time.Second
	}

	missing := 1 - b.tokens

	retry := time.Duration(
		(missing / l.refillRate) *
			float64(time.Second),
	)

	if retry < time.Millisecond {
		retry = time.Millisecond
	}

	return false, retry
}
