package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

func requestWithIdentity(
	req *http.Request,
	principal identity.Identity,
) *http.Request {
	ctx := context.WithValue(
		req.Context(),
		identityKey{},
		principal,
	)

	return req.WithContext(ctx)
}

func TestRepeated429TriggersCooldown(
	t *testing.T,
) {
	tracker := abuse.New(
		3,
		time.Minute,
		time.Hour,
	)

	handler := AbuseGuard(
		tracker,
		nil,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusTooManyRequests,
			)
		}),
	)

	principal := identity.Identity{
		UserID:  "user_1",
		KeyID:   1,
		KeyName: "key_1",
	}

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/v1/tools/evaluate",
			nil,
		)

		req = requestWithIdentity(
			req,
			principal,
		)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(
			rec,
			req,
		)
	}

	blocked, _, _ := tracker.Check(
		"api_key:1",
	)

	if !blocked {
		t.Fatal(
			"expected key to enter cooldown",
		)
	}
}

func TestSecurity403AddsThreePoints(
	t *testing.T,
) {
	tracker := abuse.New(
		5,
		time.Minute,
		time.Hour,
	)

	handler := AbuseGuard(
		tracker,
		nil,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusForbidden,
			)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/rag/chat",
		nil,
	)

	req = requestWithIdentity(
		req,
		identity.Identity{
			UserID:  "user_1",
			KeyID:   1,
			KeyName: "key_1",
		},
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	score := tracker.Score(
		"api_key:1",
	)

	if score != 3 {
		t.Fatalf(
			"expected score 3, got %d",
			score,
		)
	}
}
