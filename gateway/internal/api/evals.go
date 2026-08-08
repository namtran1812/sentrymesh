package api

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
)

// Production fallback.
//
// Local development still reads evals/results/latest.json
// so `make eval` remains the source of truth.
//
//go:embed evals_latest.json
var embeddedEvalResults []byte

func EvalResultsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	root := os.Getenv(
		"SENTRYMESH_ROOT",
	)

	if root == "" {
		root = ".."
	}

	path :=
		root +
			"/evals/results/latest.json"

	data, err := os.ReadFile(path)

	if err != nil {
		// Docker/production builds may not contain
		// the repository-level evals directory.
		data = embeddedEvalResults
	}

	var result any

	if err := json.Unmarshal(
		data,
		&result,
	); err != nil {
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"invalid eval results",
		)
		return
	}

	_ = json.NewEncoder(w).Encode(
		result,
	)
}
