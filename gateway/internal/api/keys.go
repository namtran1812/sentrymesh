package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
)

type CreateKeyRequest struct {
	Name      string   `json:"name"`
	UserID    string   `json:"user_id"`
	Role      string   `json:"role"`
	Team      string   `json:"team"`
	Scopes    []string `json:"scopes"`
	ExpiresIn int      `json:"expires_in_hours,omitempty"`
}

func ListKeysHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	keys, err := auth.DefaultStore.List(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to list API keys"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(keys)
}

func CreateKeyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	actor, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			`{"error":"authenticated identity unavailable"}`,
			http.StatusUnauthorized,
		)
		return
	}

	var req CreateKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
		return
	}

	rawKey, err := auth.GenerateRawKey()
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to generate API key"}`,
			http.StatusInternalServerError,
		)
		return
	}

	var expiresAt *time.Time

	if req.ExpiresIn > 0 {
		value := time.Now().Add(
			time.Duration(req.ExpiresIn) * time.Hour,
		)
		expiresAt = &value
	}

	principal := identity.Identity{
		UserID: req.UserID,
		Role:   identity.Role(req.Role),
		Team:   req.Team,
		Scopes: req.Scopes,
	}

	if err := auth.DefaultStore.Create(
		r.Context(),
		req.Name,
		rawKey,
		principal,
		expiresAt,
	); err != nil {
		http.Error(
			w,
			`{"error":"failed to create API key"}`,
			http.StatusInternalServerError,
		)
		return
	}

	key, err := auth.DefaultStore.FindByName(
		r.Context(),
		req.Name,
	)
	if err == nil {
		err = auditStore.WriteAuthEvent(
			r.Context(),
			audit.AuthEvent{
				Timestamp: time.Now(),
				EventType: "API_KEY_CREATED",
				KeyID:     key.ID,
				KeyName:   key.Name,
				Actor:     actor,
				Details: map[string]any{
					"user_id":    req.UserID,
					"role":       req.Role,
					"team":       req.Team,
					"scopes":     req.Scopes,
					"expires_at": key.ExpiresAt,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to write API_KEY_CREATED event: %v",
				err,
			)
		}
	}

	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"name":    req.Name,
			"api_key": rawKey,
			"warning": "this key will only be shown once",
		},
	)
}

func RevokeKeyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	actor, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			`{"error":"authenticated identity unavailable"}`,
			http.StatusUnauthorized,
		)
		return
	}

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"invalid key id"}`,
			http.StatusBadRequest,
		)
		return
	}

	key, err := auth.DefaultStore.FindByID(
		r.Context(),
		id,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"key not found"}`,
			http.StatusNotFound,
		)
		return
	}

	if err := auth.DefaultStore.RevokeByID(
		r.Context(),
		id,
	); err != nil {
		http.Error(
			w,
			`{"error":"key not found or already revoked"}`,
			http.StatusNotFound,
		)
		return
	}

	if err := auditStore.WriteAuthEvent(
		r.Context(),
		audit.AuthEvent{
			Timestamp: time.Now(),
			EventType: "API_KEY_REVOKED",
			KeyID:     key.ID,
			KeyName:   key.Name,
			Actor:     actor,
			Details: map[string]any{
				"user_id": key.UserID,
				"role":    key.Role,
				"team":    key.Team,
				"scopes":  key.Scopes,
			},
		},
	); err != nil {
		log.Printf(
			"failed to write API_KEY_REVOKED event: %v",
			err,
		)
	}

	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"id":      id,
			"revoked": true,
		},
	)
}
