package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/metrics"
)

func envPositiveInt(
	name string,
	fallback int,
) (int, error) {
	raw := os.Getenv(name)

	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf(
			"%s must be a positive integer",
			name,
		)
	}

	return value, nil
}

func configureAuditMode(
	deps *Dependencies,
) (*audit.AsyncRepository, error) {
	benchmarkMode :=
		os.Getenv(
			"SENTRYMESH_BENCHMARK_MODE",
		) == "1"

	// Never allow this diagnostic switch outside benchmark mode.
	if os.Getenv(
		"SENTRYMESH_DISABLE_AUDIT_WRITE",
	) == "1" {
		if !benchmarkMode {
			return nil, fmt.Errorf(
				"SENTRYMESH_DISABLE_AUDIT_WRITE requires benchmark mode",
			)
		}

		deps.Audit =
			audit.NewNoopRepository(
				deps.Audit,
			)

		log.Println(
			"benchmark mode: audit event writes disabled",
		)

		return nil, nil
	}

	if os.Getenv(
		"SENTRYMESH_AUDIT_FAIL_WRITES",
	) == "1" {
		if !benchmarkMode {
			return nil, fmt.Errorf(
				"SENTRYMESH_AUDIT_FAIL_WRITES requires benchmark mode",
			)
		}

		deps.Audit =
			audit.NewFailingRepository(
				deps.Audit,
				true,
			)

		log.Println(
			"benchmark mode: audit persistence failures injected",
		)
	}

	mode := os.Getenv(
		"SENTRYMESH_AUDIT_MODE",
	)

	if mode == "" || mode == "sync" {
		log.Println(
			"audit persistence mode: synchronous",
		)

		return nil, nil
	}

	if mode != "async" {
		return nil, fmt.Errorf(
			"unsupported SENTRYMESH_AUDIT_MODE %q",
			mode,
		)
	}

	queueSize, err :=
		envPositiveInt(
			"SENTRYMESH_AUDIT_QUEUE_SIZE",
			16_384,
		)
	if err != nil {
		return nil, err
	}

	batchSize, err :=
		envPositiveInt(
			"SENTRYMESH_AUDIT_BATCH_SIZE",
			128,
		)
	if err != nil {
		return nil, err
	}

	flushMS, err :=
		envPositiveInt(
			"SENTRYMESH_AUDIT_FLUSH_MS",
			10,
		)
	if err != nil {
		return nil, err
	}

	asyncRepository, err :=
		audit.NewAsyncRepository(
			deps.Audit,
			audit.AsyncOptions{
				QueueSize: queueSize,
				BatchSize: batchSize,
				FlushInterval: time.Duration(
					flushMS,
				) * time.Millisecond,
				WriteTimeout: 5 * time.Second,
			},
		)
	if err != nil {
		return nil, err
	}

	deps.Audit = asyncRepository

	metrics.SetAsyncAuditStatsProvider(
		func() metrics.AsyncAuditStats {
			stats :=
				asyncRepository.AsyncStats()

			return metrics.AsyncAuditStats{
				QueueDepth:       stats.QueueDepth,
				QueueCapacity:    stats.QueueCapacity,
				Enqueued:         stats.Enqueued,
				Flushed:          stats.Flushed,
				Saturated:        stats.Saturated,
				BatchesWritten:   stats.BatchesWritten,
				EnqueueWaitNanos: stats.EnqueueWaitNanos,
			}
		},
	)

	log.Printf(
		"audit persistence mode: async queue=%d batch=%d flush=%s",
		queueSize,
		batchSize,
		time.Duration(flushMS)*time.Millisecond,
	)

	return asyncRepository, nil
}
