package audit

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var _ BatchWriter = (*PostgresStore)(nil)

func (s *PostgresStore) WriteBatch(
	ctx context.Context,
	events []Event,
) error {
	if len(events) == 0 {
		return nil
	}

	var batch pgx.Batch

	for _, event := range events {
		secret, err :=
			marshalJSON(
				event.SecretFindings,
			)
		if err != nil {
			return fmt.Errorf(
				"marshal secret findings: %w",
				err,
			)
		}

		pii, err :=
			marshalJSON(
				event.PIIFindings,
			)
		if err != nil {
			return fmt.Errorf(
				"marshal pii findings: %w",
				err,
			)
		}

		injection, err :=
			marshalJSON(
				event.InjectionFindings,
			)
		if err != nil {
			return fmt.Errorf(
				"marshal injection findings: %w",
				err,
			)
		}

		output, err :=
			marshalJSON(
				event.OutputFindings,
			)
		if err != nil {
			return fmt.Errorf(
				"marshal output findings: %w",
				err,
			)
		}

		batch.Queue(
			`
			INSERT INTO audit_events (
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
			)
			VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, $12, $13
			)
			`,
			event.RequestID,
			event.Timestamp.UTC(),
			event.Provider,
			event.Model,
			event.Decision,
			event.RiskScore,
			event.Severity,
			event.LatencyMS,
			event.LatencyUS,
			secret,
			pii,
			injection,
			output,
		)
	}

	results := s.pool.SendBatch(
		ctx,
		&batch,
	)

	for range events {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()

			return fmt.Errorf(
				"execute audit batch: %w",
				err,
			)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf(
			"close audit batch: %w",
			err,
		)
	}

	return nil
}
