package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessHealthy(t *testing.T) {
	deps := &Dependencies{
		Backend: "sqlite",
		ready: func(
			context.Context,
		) error {
			return nil
		},
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	recorder :=
		httptest.NewRecorder()

	readinessHandler(deps)(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"status":"ready"`,
	) {
		t.Fatalf(
			"unexpected body: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"backend":"sqlite"`,
	) {
		t.Fatalf(
			"unexpected backend: %s",
			recorder.Body.String(),
		)
	}
}

func TestReadinessUnavailable(t *testing.T) {
	deps := &Dependencies{
		Backend: "postgres",
		ready: func(
			context.Context,
		) error {
			return errors.New(
				"database unavailable",
			)
		},
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	recorder :=
		httptest.NewRecorder()

	readinessHandler(deps)(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusServiceUnavailable {
		t.Fatalf(
			"expected 503, got %d",
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"status":"not_ready"`,
	) {
		t.Fatalf(
			"unexpected body: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"backend":"postgres"`,
	) {
		t.Fatalf(
			"unexpected backend: %s",
			recorder.Body.String(),
		)
	}
}
