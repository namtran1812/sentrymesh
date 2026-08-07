package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/runtime"
)

type SecurityPostureKey struct {
	KeyID         int64   `json:"key_id"`
	Name          string  `json:"name"`
	UserID        string  `json:"user_id"`
	Role          string  `json:"role"`
	Team          string  `json:"team"`
	Scopes        string  `json:"scopes"`
	AbuseScore    int     `json:"abuse_score"`
	Status        string  `json:"status"`
	CooldownUntil *string `json:"cooldown_until,omitempty"`
	RevokedAt     *string `json:"revoked_at,omitempty"`
}

type SecurityPostureSummary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Elevated int `json:"elevated"`
	Cooldown int `json:"cooldown"`
	Revoked  int `json:"revoked"`
}

type SecurityPostureResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Summary   SecurityPostureSummary `json:"summary"`
	Keys      []SecurityPostureKey   `json:"keys"`
}

func SecurityPostureHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	keys, err := auth.DefaultStore.List(
		r.Context(),
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to list API keys"}`,
			http.StatusInternalServerError,
		)
		return
	}

	states, err := runtime.AbuseStore.List(
		r.Context(),
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to load abuse state"}`,
			http.StatusInternalServerError,
		)
		return
	}

	stateByKey := make(map[int64]int)
	cooldownByKey := make(map[int64]*time.Time)

	for _, state := range states {
		stateByKey[state.KeyID] = state.Score
		cooldownByKey[state.KeyID] =
			state.CooldownUntil
	}

	response := SecurityPostureResponse{
		Timestamp: time.Now().UTC(),
		Keys:      make([]SecurityPostureKey, 0, len(keys)),
	}

	for _, key := range keys {
		score := stateByKey[key.ID]
		cooldown := cooldownByKey[key.ID]

		status := "HEALTHY"

		if key.RevokedAt != nil {
			status = "REVOKED"
			response.Summary.Revoked++
		} else if cooldown != nil &&
			time.Now().Before(*cooldown) {

			status = "COOLDOWN"
			response.Summary.Cooldown++
		} else if score > 0 {
			status = "ELEVATED"
			response.Summary.Elevated++
		} else {
			response.Summary.Healthy++
		}

		var cooldownString *string

		if cooldown != nil {
			value := cooldown.UTC().
				Format(time.RFC3339Nano)

			cooldownString = &value
		}

		response.Keys = append(
			response.Keys,
			SecurityPostureKey{
				KeyID:         key.ID,
				Name:          key.Name,
				UserID:        key.UserID,
				Role:          key.Role,
				Team:          key.Team,
				Scopes:        key.Scopes,
				AbuseScore:    score,
				Status:        status,
				CooldownUntil: cooldownString,
				RevokedAt:     key.RevokedAt,
			},
		)
	}

	response.Summary.Total = len(response.Keys)

	_ = json.NewEncoder(w).Encode(response)
}
