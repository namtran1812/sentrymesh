package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
)

func newRequestID() string {
	buf := make([]byte, 8)

	if _, err := rand.Read(buf); err != nil {
		return "req_unknown"
	}

	return "req_" + hex.EncodeToString(buf)
}

func requestID(r *http.Request) string {
	if value := middleware.RequestIDFromContext(
		r.Context(),
	); value != "" {
		return value
	}

	return newRequestID()
}
