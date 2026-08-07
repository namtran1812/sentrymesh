package abuse

import (
	"context"
	"sync"
	"time"
)

type state struct {
	score         int
	lastUpdated   time.Time
	cooldownUntil time.Time
}

type Tracker struct {
	mu sync.Mutex

	threshold  int
	cooldown   time.Duration
	decayEvery time.Duration
	states     map[string]*state
	keyIDs     map[string]int64

	now   func() time.Time
	store *Store
}

func New(
	threshold int,
	cooldown time.Duration,
	decayEvery time.Duration,
) *Tracker {
	return &Tracker{
		threshold:  threshold,
		cooldown:   cooldown,
		decayEvery: decayEvery,
		states:     make(map[string]*state),
		keyIDs:     make(map[string]int64),
		now:        time.Now,
	}
}

func NewPersistent(
	threshold int,
	cooldown time.Duration,
	decayEvery time.Duration,
	store *Store,
) *Tracker {
	tracker := New(
		threshold,
		cooldown,
		decayEvery,
	)

	tracker.store = store

	return tracker
}

func (t *Tracker) RegisterKey(
	ctx context.Context,
	key string,
	keyID int64,
) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.keyIDs[key] = keyID

	if t.store == nil {
		return
	}

	stored, err := t.store.Load(
		ctx,
		keyID,
	)
	if err != nil {
		return
	}

	s := &state{
		score:       stored.Score,
		lastUpdated: stored.LastUpdated,
	}

	if stored.CooldownUntil != nil {
		s.cooldownUntil =
			*stored.CooldownUntil
	}

	t.states[key] = s
}

func (t *Tracker) applyDecay(
	s *state,
	now time.Time,
) {
	if t.decayEvery <= 0 {
		return
	}

	elapsed := now.Sub(s.lastUpdated)

	if elapsed < t.decayEvery {
		return
	}

	steps := int(
		elapsed /
			t.decayEvery,
	)

	s.score -= steps

	if s.score < 0 {
		s.score = 0
	}

	s.lastUpdated = s.lastUpdated.Add(
		time.Duration(steps) *
			t.decayEvery,
	)
}

func (t *Tracker) persist(
	key string,
	s *state,
) {
	if t.store == nil {
		return
	}

	keyID, ok := t.keyIDs[key]
	if !ok {
		return
	}

	var cooldown *time.Time

	if !s.cooldownUntil.IsZero() {
		value := s.cooldownUntil
		cooldown = &value
	}

	_ = t.store.Save(
		context.Background(),
		StoredState{
			KeyID:         keyID,
			Score:         s.score,
			LastUpdated:   s.lastUpdated,
			CooldownUntil: cooldown,
		},
	)
}

func (t *Tracker) Add(
	key string,
	points int,
) (
	score int,
	cooldownUntil time.Time,
	enteredCooldown bool,
) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()

	s, ok := t.states[key]
	if !ok {
		s = &state{
			lastUpdated: now,
		}

		t.states[key] = s
	}

	t.applyDecay(s, now)

	s.score += points
	s.lastUpdated = now

	if s.score >= t.threshold &&
		!now.Before(
			s.cooldownUntil,
		) {

		s.cooldownUntil =
			now.Add(t.cooldown)

		t.persist(key, s)

		return s.score,
			s.cooldownUntil,
			true
	}

	t.persist(key, s)

	return s.score,
		s.cooldownUntil,
		false
}

func (t *Tracker) Check(
	key string,
) (
	blocked bool,
	retryAfter time.Duration,
	score int,
) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()

	s, ok := t.states[key]
	if !ok {
		return false, 0, 0
	}

	t.applyDecay(s, now)

	if now.Before(
		s.cooldownUntil,
	) {
		t.persist(key, s)

		return true,
			s.cooldownUntil.Sub(now),
			s.score
	}

	if !s.cooldownUntil.IsZero() {
		s.cooldownUntil =
			time.Time{}
	}

	t.persist(key, s)

	return false, 0, s.score
}

func (t *Tracker) Score(
	key string,
) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, ok := t.states[key]
	if !ok {
		return 0
	}

	now := t.now()

	t.applyDecay(s, now)
	t.persist(key, s)

	return s.score
}
