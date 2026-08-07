package abuse

import (
	"sync"
	"testing"
	"time"
)

func TestScoreTriggersCooldown(t *testing.T) {
	tracker := New(
		5,
		30*time.Second,
		time.Minute,
	)

	now := time.Now()

	tracker.now = func() time.Time {
		return now
	}

	for i := 0; i < 4; i++ {
		_, _, entered := tracker.Add(
			"key_1",
			1,
		)

		if entered {
			t.Fatal("cooldown triggered too early")
		}
	}

	score, _, entered := tracker.Add(
		"key_1",
		1,
	)

	if !entered {
		t.Fatal("expected cooldown")
	}

	if score != 5 {
		t.Fatalf(
			"expected score 5, got %d",
			score,
		)
	}

	blocked, _, _ := tracker.Check("key_1")

	if !blocked {
		t.Fatal("expected key to be blocked")
	}
}

func TestCooldownExpires(t *testing.T) {
	tracker := New(
		2,
		10*time.Second,
		time.Minute,
	)

	now := time.Now()

	tracker.now = func() time.Time {
		return now
	}

	tracker.Add("key_1", 2)

	blocked, _, _ := tracker.Check("key_1")

	if !blocked {
		t.Fatal("expected active cooldown")
	}

	now = now.Add(11 * time.Second)

	blocked, _, _ = tracker.Check("key_1")

	if blocked {
		t.Fatal("expected cooldown to expire")
	}
}

func TestScoreDecays(t *testing.T) {
	tracker := New(
		10,
		time.Minute,
		10*time.Second,
	)

	now := time.Now()

	tracker.now = func() time.Time {
		return now
	}

	tracker.Add("key_1", 5)

	now = now.Add(20 * time.Second)

	score := tracker.Score("key_1")

	if score != 3 {
		t.Fatalf(
			"expected score 3 after decay, got %d",
			score,
		)
	}
}

func TestKeysAreIndependent(t *testing.T) {
	tracker := New(
		3,
		time.Minute,
		time.Minute,
	)

	tracker.Add("key_a", 3)

	blockedA, _, _ := tracker.Check("key_a")
	blockedB, _, _ := tracker.Check("key_b")

	if !blockedA {
		t.Fatal("key_a should be blocked")
	}

	if blockedB {
		t.Fatal("key_b should not be blocked")
	}
}

func TestConcurrentScoreUpdates(t *testing.T) {
	tracker := New(
		1000,
		time.Minute,
		time.Hour,
	)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			tracker.Add(
				"shared_key",
				1,
			)
		}()
	}

	wg.Wait()

	score := tracker.Score("shared_key")

	if score != 100 {
		t.Fatalf(
			"expected score 100, got %d",
			score,
		)
	}
}
