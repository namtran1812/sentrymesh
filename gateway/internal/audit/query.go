package audit

import (
	"context"
	"encoding/json"
	"fmt"
)

type Record struct {
	ID                int64           `json:"id"`
	RequestID         string          `json:"request_id"`
	Timestamp         string          `json:"timestamp"`
	Provider          string          `json:"provider"`
	Model             string          `json:"model"`
	Decision          string          `json:"decision"`
	RiskScore         int             `json:"risk_score"`
	Severity          string          `json:"severity"`
	LatencyMS         int64           `json:"latency_ms"`
	LatencyUS         int64           `json:"latency_us"`
	SecretFindings    json.RawMessage `json:"secret_findings"`
	PIIFindings       json.RawMessage `json:"pii_findings"`
	InjectionFindings json.RawMessage `json:"injection_findings"`
	OutputFindings    json.RawMessage `json:"output_findings"`
}

func (s *Store) List(ctx context.Context, limit int) ([]Record, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			request_id,
			timestamp,
			provider,
			model,
			decision,
			risk_score,
			severity,
			latency_ms,
			latency_us,
			secret_findings,
			pii_findings,
			injection_findings,
			output_findings
		FROM audit_events
		ORDER BY id DESC
		LIMIT ?
		`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	records := make([]Record, 0)

	for rows.Next() {
		var record Record
		var secret string
		var pii string
		var injection string
		var output string

		if err := rows.Scan(
			&record.ID,
			&record.RequestID,
			&record.Timestamp,
			&record.Provider,
			&record.Model,
			&record.Decision,
			&record.RiskScore,
			&record.Severity,
			&record.LatencyMS,
			&record.LatencyUS,
			&secret,
			&pii,
			&injection,
			&output,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}

		record.SecretFindings = json.RawMessage(secret)
		record.PIIFindings = json.RawMessage(pii)
		record.InjectionFindings = json.RawMessage(injection)
		record.OutputFindings = json.RawMessage(output)

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}

	return records, nil
}
