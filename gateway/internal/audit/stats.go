package audit

import (
	"context"
	"fmt"
)

type Stats struct {
	TotalRequests        int     `json:"total_requests"`
	AllowedRequests      int     `json:"allowed_requests"`
	RedactedRequests     int     `json:"redacted_requests"`
	BlockedRequests      int     `json:"blocked_requests"`
	AverageLatencyMS     float64 `json:"average_latency_ms"`
	AverageRiskScore     float64 `json:"average_risk_score"`
	CriticalEvents       int     `json:"critical_events"`
	HighSeverityEvents   int     `json:"high_severity_events"`
	MediumSeverityEvents int     `json:"medium_severity_events"`
}

func (s *Store) Stats(
	ctx context.Context,
) (Stats, error) {
	var stats Stats

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			COUNT(*),

			COALESCE(SUM(
				CASE
					WHEN decision = 'ALLOW' THEN 1
					ELSE 0
				END
			), 0),

			COALESCE(SUM(
				CASE
					WHEN decision = 'ALLOW_WITH_REDACTION' THEN 1
					ELSE 0
				END
			), 0),

			COALESCE(SUM(
				CASE
					WHEN decision = 'BLOCK' THEN 1
					ELSE 0
				END
			), 0),

			COALESCE(AVG(latency_ms), 0),
			COALESCE(AVG(risk_score), 0),

			COALESCE(SUM(
				CASE
					WHEN severity = 'CRITICAL' THEN 1
					ELSE 0
				END
			), 0),

			COALESCE(SUM(
				CASE
					WHEN severity = 'HIGH' THEN 1
					ELSE 0
				END
			), 0),

			COALESCE(SUM(
				CASE
					WHEN severity = 'MEDIUM' THEN 1
					ELSE 0
				END
			), 0)

		FROM audit_events
		`,
	).Scan(
		&stats.TotalRequests,
		&stats.AllowedRequests,
		&stats.RedactedRequests,
		&stats.BlockedRequests,
		&stats.AverageLatencyMS,
		&stats.AverageRiskScore,
		&stats.CriticalEvents,
		&stats.HighSeverityEvents,
		&stats.MediumSeverityEvents,
	)

	if err != nil {
		return Stats{},
			fmt.Errorf(
				"query audit stats: %w",
				err,
			)
	}

	return stats, nil
}
