package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var requestTotal atomic.Uint64
var requestDurationUS atomic.Uint64

var securityBlocks atomic.Uint64
var redactions atomic.Uint64
var rateLimits atomic.Uint64
var abuseCooldowns atomic.Uint64
var providerErrors atomic.Uint64

var durationBuckets = [...]struct {
	upper uint64
	count atomic.Uint64
}{
	{upper: 100},
	{upper: 250},
	{upper: 500},
	{upper: 1000},
	{upper: 2500},
	{upper: 5000},
	{upper: 10000},
	{upper: 25000},
	{upper: 50000},
	{upper: 100000},
	{upper: 250000},
	{upper: 500000},
	{upper: 1000000},
}

func ObserveRequest(duration time.Duration) {
	us := uint64(duration.Microseconds())

	requestTotal.Add(1)
	requestDurationUS.Add(us)

	for i := range durationBuckets {
		if us <= durationBuckets[i].upper {
			durationBuckets[i].count.Add(1)
		}
	}
}

func IncSecurityBlock() {
	securityBlocks.Add(1)
}

func IncRedaction() {
	redactions.Add(1)
}

func IncRateLimit() {
	rateLimits.Add(1)
}

func IncAbuseCooldown() {
	abuseCooldowns.Add(1)
}

func IncProviderError() {
	providerErrors.Add(1)
}

func Handler() http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.Header().Set(
			"Content-Type",
			"text/plain; version=0.0.4; charset=utf-8",
		)

		total := requestTotal.Load()

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_requests_total Total HTTP requests processed.\n"+
				"# TYPE sentrymesh_requests_total counter\n"+
				"sentrymesh_requests_total %d\n",
			total,
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_request_duration_microseconds HTTP request latency in microseconds.\n"+
				"# TYPE sentrymesh_request_duration_microseconds histogram\n",
		)

		for i := range durationBuckets {
			fmt.Fprintf(
				w,
				"sentrymesh_request_duration_microseconds_bucket{le=\"%d\"} %d\n",
				durationBuckets[i].upper,
				durationBuckets[i].count.Load(),
			)
		}

		fmt.Fprintf(
			w,
			"sentrymesh_request_duration_microseconds_bucket{le=\"+Inf\"} %d\n",
			total,
		)

		fmt.Fprintf(
			w,
			"sentrymesh_request_duration_microseconds_sum %d\n",
			requestDurationUS.Load(),
		)

		fmt.Fprintf(
			w,
			"sentrymesh_request_duration_microseconds_count %d\n",
			total,
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_security_blocks_total Requests blocked by security policy.\n"+
				"# TYPE sentrymesh_security_blocks_total counter\n"+
				"sentrymesh_security_blocks_total %d\n",
			securityBlocks.Load(),
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_redactions_total Requests allowed with sensitive-data redaction.\n"+
				"# TYPE sentrymesh_redactions_total counter\n"+
				"sentrymesh_redactions_total %d\n",
			redactions.Load(),
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_rate_limits_total Requests rejected by rate limiting.\n"+
				"# TYPE sentrymesh_rate_limits_total counter\n"+
				"sentrymesh_rate_limits_total %d\n",
			rateLimits.Load(),
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_abuse_cooldowns_total API keys placed into abuse cooldown.\n"+
				"# TYPE sentrymesh_abuse_cooldowns_total counter\n"+
				"sentrymesh_abuse_cooldowns_total %d\n",
			abuseCooldowns.Load(),
		)

		fmt.Fprintf(
			w,
			"# HELP sentrymesh_provider_errors_total Model provider failures.\n"+
				"# TYPE sentrymesh_provider_errors_total counter\n"+
				"sentrymesh_provider_errors_total %d\n",
			providerErrors.Load(),
		)
	})
}
