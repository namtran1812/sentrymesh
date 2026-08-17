package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	Concurrency int               `json:"concurrency"`
	Requests    int               `json:"requests"`
	Successes   uint64            `json:"successes"`
	Errors      uint64            `json:"errors"`
	DurationMS  float64           `json:"duration_ms"`
	Throughput  float64           `json:"requests_per_second"`
	P50US       int64             `json:"p50_us"`
	P95US       int64             `json:"p95_us"`
	P99US       int64             `json:"p99_us"`
	P999US      int64             `json:"p999_us"`
	MaxUS       int64             `json:"max_us"`
	StatusCodes map[int]uint64    `json:"status_codes"`
	Decisions   map[string]uint64 `json:"decisions"`
}

type response struct {
	Decision string `json:"decision"`
}

func percentile(values []int64, p float64) int64 {
	if len(values) == 0 {
		return 0
	}

	index := int(float64(len(values)-1) * p)
	return values[index]
}

func run(
	client *http.Client,
	url string,
	token string,
	concurrency int,
	requests int,
) result {
	payload := []byte(`{
		"model":"benchmark",
		"provider":"mock",
		"messages":[
			{
				"role":"user",
				"content":"Summarize this benchmark report."
			}
		]
	}`)

	jobs := make(chan struct{}, requests)
	latencies := make([]int64, 0, requests)

	var latencyMu sync.Mutex
	var statusMu sync.Mutex
	var decisionMu sync.Mutex

	statusCodes := map[int]uint64{}
	decisions := map[string]uint64{}

	var successes atomic.Uint64
	var errors atomic.Uint64

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()

		for range jobs {
			req, err := http.NewRequest(
				http.MethodPost,
				url,
				bytes.NewReader(payload),
			)
			if err != nil {
				errors.Add(1)
				continue
			}

			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			start := time.Now()

			res, err := client.Do(req)

			elapsed := time.Since(start).Microseconds()

			latencyMu.Lock()
			latencies = append(latencies, elapsed)
			latencyMu.Unlock()

			if err != nil {
				errors.Add(1)
				continue
			}

			body, err := io.ReadAll(res.Body)
			res.Body.Close()

			statusMu.Lock()
			statusCodes[res.StatusCode]++
			statusMu.Unlock()

			if err != nil {
				errors.Add(1)
				continue
			}

			var decoded response
			if err := json.Unmarshal(body, &decoded); err == nil &&
				decoded.Decision != "" {
				decisionMu.Lock()
				decisions[decoded.Decision]++
				decisionMu.Unlock()
			}

			if res.StatusCode >= 200 && res.StatusCode < 300 {
				successes.Add(1)
			} else {
				errors.Add(1)
			}
		}
	}

	started := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	for i := 0; i < requests; i++ {
		jobs <- struct{}{}
	}

	close(jobs)
	wg.Wait()

	duration := time.Since(started)

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var max int64
	if len(latencies) > 0 {
		max = latencies[len(latencies)-1]
	}

	return result{
		Concurrency: concurrency,
		Requests:    requests,
		Successes:   successes.Load(),
		Errors:      errors.Load(),
		DurationMS:  float64(duration.Microseconds()) / 1000,
		Throughput:  float64(requests) / duration.Seconds(),
		P50US:       percentile(latencies, 0.50),
		P95US:       percentile(latencies, 0.95),
		P99US:       percentile(latencies, 0.99),
		P999US:      percentile(latencies, 0.999),
		MaxUS:       max,
		StatusCodes: statusCodes,
		Decisions:   decisions,
	}
}

func parseLevels(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	levels := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf(
				"invalid concurrency level %q",
				part,
			)
		}

		levels = append(levels, value)
	}

	if len(levels) == 0 {
		return nil, fmt.Errorf(
			"at least one concurrency level is required",
		)
	}

	return levels, nil
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	mid := len(sorted) / 2

	if len(sorted)%2 == 1 {
		return sorted[mid]
	}

	return (sorted[mid-1] + sorted[mid]) / 2
}

func medianInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	mid := len(sorted) / 2

	if len(sorted)%2 == 1 {
		return sorted[mid]
	}

	return (sorted[mid-1] + sorted[mid]) / 2
}

func main() {
	var (
		baseURL = flag.String(
			"url",
			"http://127.0.0.1:8080/v1/chat/completions",
			"benchmark endpoint",
		)
		token = flag.String(
			"token",
			"sm_admin_dev",
			"API token",
		)
		requests = flag.Int(
			"requests",
			5000,
			"requests per concurrency level",
		)
		warmup = flag.Int(
			"warmup",
			500,
			"warmup requests",
		)
		output = flag.String(
			"output",
			"",
			"optional JSON output path",
		)
		levelsRaw = flag.String(
			"levels",
			"1,2,4,8,16,32,64",
			"comma-separated concurrency levels",
		)
		repeat = flag.Int(
			"repeat",
			1,
			"number of repetitions per concurrency level",
		)
	)

	flag.Parse()

	if *repeat <= 0 {
		panic("repeat must be positive")
	}

	concurrencyLevels, err := parseLevels(
		*levelsRaw,
	)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1024,
			MaxIdleConnsPerHost: 1024,
			MaxConnsPerHost:     1024,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	fmt.Printf("warming up with %d requests...\n", *warmup)

	_ = run(
		client,
		*baseURL,
		*token,
		8,
		*warmup,
	)

	results := make([]result, 0, len(concurrencyLevels)*(*repeat))

	fmt.Println()
	fmt.Printf(
		"%-6s %-6s %-10s %-10s %-10s %-10s %-10s %-10s\n",
		"CONC",
		"RUN",
		"REQ/S",
		"P50(us)",
		"P95(us)",
		"P99(us)",
		"P999(us)",
		"ERRORS",
	)

	for _, concurrency := range concurrencyLevels {
		rpsSamples := make([]float64, 0, *repeat)
		p50Samples := make([]int64, 0, *repeat)
		p95Samples := make([]int64, 0, *repeat)
		p99Samples := make([]int64, 0, *repeat)
		p999Samples := make([]int64, 0, *repeat)

		for runIndex := 1; runIndex <= *repeat; runIndex++ {
			r := run(
				client,
				*baseURL,
				*token,
				concurrency,
				*requests,
			)

			results = append(results, r)

			rpsSamples = append(
				rpsSamples,
				r.Throughput,
			)
			p50Samples = append(
				p50Samples,
				r.P50US,
			)
			p95Samples = append(
				p95Samples,
				r.P95US,
			)
			p99Samples = append(
				p99Samples,
				r.P99US,
			)
			p999Samples = append(
				p999Samples,
				r.P999US,
			)

			fmt.Printf(
				"%-6d %-6d %-10.1f %-10d %-10d %-10d %-10d %-10d\n",
				r.Concurrency,
				runIndex,
				r.Throughput,
				r.P50US,
				r.P95US,
				r.P99US,
				r.P999US,
				r.Errors,
			)
		}

		if *repeat > 1 {
			fmt.Printf(
				"%-6d %-6s %-10.1f %-10d %-10d %-10d %-10d %-10s\n",
				concurrency,
				"MED",
				medianFloat(rpsSamples),
				medianInt64(p50Samples),
				medianInt64(p95Samples),
				medianInt64(p99Samples),
				medianInt64(p999Samples),
				"-",
			)
		}
	}

	if *output == "" {
		return
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(*output, data, 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("\nwrote %s\n", *output)
}
