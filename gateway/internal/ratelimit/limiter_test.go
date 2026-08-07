package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestAllowsInitialBurst(t *testing.T) {
	limiter := New(3, 1)

	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow("key_1")

		if !allowed {
			t.Fatalf(
				"request %d should be allowed",
				i+1,
			)
		}
	}

	allowed, _ := limiter.Allow("key_1")

	if allowed {
		t.Fatal("fourth request should be rate limited")
	}
}

func TestBucketsAreIndependent(t *testing.T) {
	limiter := New(1, 1)

	allowed, _ := limiter.Allow("key_a")
	if !allowed {
		t.Fatal("key_a should be allowed")
	}

	allowed, _ = limiter.Allow("key_a")
	if allowed {
		t.Fatal("key_a should be limited")
	}

	allowed, _ = limiter.Allow("key_b")
	if !allowed {
		t.Fatal("key_b should have its own bucket")
	}
}

func TestTokensRefill(t *testing.T) {
	limiter := New(1, 1)

	now := time.Now()

	limiter.now = func() time.Time {
		return now
	}

	allowed, _ := limiter.Allow("key_1")
	if !allowed {
		t.Fatal("first request should be allowed")
	}

	allowed, _ = limiter.Allow("key_1")
	if allowed {
		t.Fatal("second request should be limited")
	}

	now = now.Add(time.Second)

	allowed, _ = limiter.Allow("key_1")
	if !allowed {
		t.Fatal("token should refill after one second")
	}
}

func TestConcurrentAccessDoesNotExceedCapacity(
	t *testing.T,
) {
	limiter := New(10, 0)

	var wg sync.WaitGroup
	var mu sync.Mutex

	allowedCount := 0

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			allowed, _ := limiter.Allow(
				"shared_key",
			)

			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != 10 {
		t.Fatalf(
			"expected 10 allowed requests, got %d",
			allowedCount,
		)
	}
}
