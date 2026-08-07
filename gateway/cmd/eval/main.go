package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
	"github.com/namtran1812/sentrymesh/gateway/internal/rag"
	"github.com/namtran1812/sentrymesh/gateway/internal/risk"
	"github.com/namtran1812/sentrymesh/gateway/internal/scanner"
)

type InjectionCase struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type PIICase struct {
	Name               string `json:"name"`
	Input              string `json:"input"`
	ExpectedRedactions int    `json:"expected_redactions"`
}

type RAGCase struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Team string `json:"team"`

	Documents []rag.Document `json:"documents"`

	ExpectedIncluded int `json:"expected_included"`
}

type Metrics struct {
	Total  int
	Passed int
	Failed int

	TruePositive  int
	TrueNegative  int
	FalsePositive int
	FalseNegative int

	TotalLatency time.Duration
}

type EvalMetricReport struct {
	Total         int     `json:"total"`
	Passed        int     `json:"passed"`
	Failed        int     `json:"failed"`
	Accuracy      float64 `json:"accuracy"`
	Precision     float64 `json:"precision,omitempty"`
	Recall        float64 `json:"recall,omitempty"`
	FalsePositive int     `json:"false_positives,omitempty"`
	FalseNegative int     `json:"false_negatives,omitempty"`
	AverageNS     int64   `json:"average_latency_ns"`
}

type EvalReport struct {
	Timestamp       time.Time        `json:"timestamp"`
	PromptInjection EvalMetricReport `json:"prompt_injection"`
	PII             EvalMetricReport `json:"pii"`
	RAG             EvalMetricReport `json:"rag"`
}

func evalPath(name string) string {
	root := os.Getenv("SENTRYMESH_ROOT")

	if root == "" {
		root = ".."
	}

	return filepath.Join(
		root,
		"evals",
		"cases",
		name,
	)
}

func loadJSON[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result []T

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func runInjection() Metrics {
	cases, err := loadJSON[InjectionCase](
		evalPath("injection.json"),
	)
	if err != nil {
		panic(err)
	}

	var m Metrics

	fmt.Println()
	fmt.Println("PROMPT INJECTION")
	fmt.Println("----------------")

	for _, c := range cases {
		start := time.Now()

		findings := scanner.ScanPromptInjection(c.Input)

		maxConfidence := 0

		for _, finding := range findings {
			if finding.Confidence > maxConfidence {
				maxConfidence = finding.Confidence
			}
		}

		decision := risk.Evaluate(
			risk.Input{
				InjectionCount: len(findings),
				MaxInjection:   maxConfidence,
			},
		)

		elapsed := time.Since(start)

		actual := decision.Action

		if actual == "REVIEW" {
			actual = "BLOCK"
		}

		expectedMalicious := c.Expected == "BLOCK"
		actualMalicious := actual == "BLOCK"

		switch {
		case expectedMalicious && actualMalicious:
			m.TruePositive++

		case !expectedMalicious && !actualMalicious:
			m.TrueNegative++

		case !expectedMalicious && actualMalicious:
			m.FalsePositive++

		case expectedMalicious && !actualMalicious:
			m.FalseNegative++
		}

		m.Total++
		m.TotalLatency += elapsed

		if actual == c.Expected {
			m.Passed++

			fmt.Printf(
				"PASS %-32s %s\n",
				c.Name,
				actual,
			)
		} else {
			m.Failed++

			fmt.Printf(
				"FAIL %-32s expected=%s actual=%s\n",
				c.Name,
				c.Expected,
				actual,
			)
		}
	}

	return m
}

func runPII() Metrics {
	cases, err := loadJSON[PIICase](
		evalPath("pii.json"),
	)
	if err != nil {
		panic(err)
	}

	var m Metrics

	fmt.Println()
	fmt.Println("PII REDACTION")
	fmt.Println("-------------")

	for _, c := range cases {
		start := time.Now()

		_, findings := scanner.ScanAndRedactPII(c.Input)

		elapsed := time.Since(start)

		m.Total++
		m.TotalLatency += elapsed

		if len(findings) == c.ExpectedRedactions {
			m.Passed++

			fmt.Printf(
				"PASS %-32s findings=%d\n",
				c.Name,
				len(findings),
			)
		} else {
			m.Failed++

			fmt.Printf(
				"FAIL %-32s expected=%d actual=%d\n",
				c.Name,
				c.ExpectedRedactions,
				len(findings),
			)
		}
	}

	return m
}

func runRAG() Metrics {
	cases, err := loadJSON[RAGCase](
		evalPath("rag.json"),
	)
	if err != nil {
		panic(err)
	}

	var m Metrics

	fmt.Println()
	fmt.Println("RAG SECURITY")
	fmt.Println("------------")

	for _, c := range cases {
		start := time.Now()

		result := rag.BuildContext(
			"eval",
			identity.Identity{
				UserID: "eval-user",
				Role:   identity.Role(c.Role),
				Team:   c.Team,
			},
			c.Documents,
		)

		elapsed := time.Since(start)

		actualIncluded := len(result.Context)

		m.Total++
		m.TotalLatency += elapsed

		if actualIncluded == c.ExpectedIncluded {
			m.Passed++

			fmt.Printf(
				"PASS %-32s included=%d\n",
				c.Name,
				actualIncluded,
			)
		} else {
			m.Failed++

			fmt.Printf(
				"FAIL %-32s expected=%d actual=%d\n",
				c.Name,
				c.ExpectedIncluded,
				actualIncluded,
			)
		}
	}

	return m
}

func printMetrics(
	name string,
	m Metrics,
) {
	fmt.Println()
	fmt.Println(name)
	fmt.Println("==============================")

	accuracy := 0.0

	if m.Total > 0 {
		accuracy =
			float64(m.Passed) /
				float64(m.Total) *
				100
	}

	avgLatency := time.Duration(0)

	if m.Total > 0 {
		avgLatency =
			m.TotalLatency /
				time.Duration(m.Total)
	}

	fmt.Printf("Total:            %d\n", m.Total)
	fmt.Printf("Passed:           %d\n", m.Passed)
	fmt.Printf("Failed:           %d\n", m.Failed)
	fmt.Printf("Accuracy:         %.2f%%\n", accuracy)
	fmt.Printf("Average latency:  %s\n", avgLatency)

	if m.TruePositive+
		m.TrueNegative+
		m.FalsePositive+
		m.FalseNegative > 0 {

		fmt.Printf(
			"True positives:   %d\n",
			m.TruePositive,
		)

		fmt.Printf(
			"True negatives:   %d\n",
			m.TrueNegative,
		)

		fmt.Printf(
			"False positives:  %d\n",
			m.FalsePositive,
		)

		fmt.Printf(
			"False negatives:  %d\n",
			m.FalseNegative,
		)

		precision := 0.0
		recall := 0.0

		if m.TruePositive+m.FalsePositive > 0 {
			precision =
				float64(m.TruePositive) /
					float64(
						m.TruePositive+
							m.FalsePositive,
					)
		}

		if m.TruePositive+m.FalseNegative > 0 {
			recall =
				float64(m.TruePositive) /
					float64(
						m.TruePositive+
							m.FalseNegative,
					)
		}

		fmt.Printf(
			"Precision:        %.3f\n",
			precision,
		)

		fmt.Printf(
			"Recall:           %.3f\n",
			recall,
		)
	}
}

func metricReport(m Metrics) EvalMetricReport {
	accuracy := 0.0
	precision := 0.0
	recall := 0.0
	average := int64(0)

	if m.Total > 0 {
		accuracy =
			float64(m.Passed) /
				float64(m.Total)

		average = int64(
			m.TotalLatency /
				time.Duration(m.Total),
		)
	}

	if m.TruePositive+m.FalsePositive > 0 {
		precision =
			float64(m.TruePositive) /
				float64(
					m.TruePositive+
						m.FalsePositive,
				)
	}

	if m.TruePositive+m.FalseNegative > 0 {
		recall =
			float64(m.TruePositive) /
				float64(
					m.TruePositive+
						m.FalseNegative,
				)
	}

	return EvalMetricReport{
		Total:         m.Total,
		Passed:        m.Passed,
		Failed:        m.Failed,
		Accuracy:      accuracy,
		Precision:     precision,
		Recall:        recall,
		FalsePositive: m.FalsePositive,
		FalseNegative: m.FalseNegative,
		AverageNS:     average,
	}
}

func writeReport(
	injection Metrics,
	pii Metrics,
	ragMetrics Metrics,
) {
	report := EvalReport{
		Timestamp:       time.Now().UTC(),
		PromptInjection: metricReport(injection),
		PII:             metricReport(pii),
		RAG:             metricReport(ragMetrics),
	}

	reportData, err := json.MarshalIndent(
		report,
		"",
		"  ",
	)
	if err != nil {
		panic(err)
	}

	root := os.Getenv("SENTRYMESH_ROOT")

	if root == "" {
		root = ".."
	}

	resultDir := filepath.Join(
		root,
		"evals",
		"results",
	)

	if err := os.MkdirAll(
		resultDir,
		0755,
	); err != nil {
		panic(err)
	}

	resultPath := filepath.Join(
		resultDir,
		"latest.json",
	)

	if err := os.WriteFile(
		resultPath,
		reportData,
		0644,
	); err != nil {
		panic(err)
	}

	fmt.Printf(
		"\nReport: %s\n",
		resultPath,
	)
}

func main() {
	injection := runInjection()
	pii := runPII()
	ragMetrics := runRAG()

	printMetrics(
		"PROMPT INJECTION METRICS",
		injection,
	)

	printMetrics(
		"PII METRICS",
		pii,
	)

	printMetrics(
		"RAG METRICS",
		ragMetrics,
	)

	writeReport(
		injection,
		pii,
		ragMetrics,
	)

	if injection.Failed+
		pii.Failed+
		ragMetrics.Failed > 0 {

		os.Exit(1)
	}
}
