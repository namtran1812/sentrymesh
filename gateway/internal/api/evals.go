package api

import (
	"encoding/json"
	"net/http"
	"os"
)

func EvalResultsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	root := os.Getenv("SENTRYMESH_ROOT")
	if root == "" {
		root = ".."
	}

	path := root + "/evals/results/latest.json"

	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(
			w,
			`{"error":"eval results unavailable"}`,
			http.StatusNotFound,
		)
		return
	}

	var result any

	if err := json.Unmarshal(data, &result); err != nil {
		http.Error(
			w,
			`{"error":"invalid eval results"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}
