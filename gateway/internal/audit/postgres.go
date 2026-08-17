package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

var _ Repository = (*PostgresStore)(nil)

func marshalJSON(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func rawJSON(data []byte) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage("null")
	}

	return json.RawMessage(data)
}

func (s *PostgresStore) Write(
	ctx context.Context,
	event Event,
) error {
	secret, err := marshalJSON(event.SecretFindings)
	if err != nil {
		return fmt.Errorf("marshal secret findings: %w", err)
	}

	pii, err := marshalJSON(event.PIIFindings)
	if err != nil {
		return fmt.Errorf("marshal pii findings: %w", err)
	}

	injection, err := marshalJSON(event.InjectionFindings)
	if err != nil {
		return fmt.Errorf("marshal injection findings: %w", err)
	}

	output, err := marshalJSON(event.OutputFindings)
	if err != nil {
		return fmt.Errorf("marshal output findings: %w", err)
	}

	_, err = s.pool.Exec(
		ctx,
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
			secret_findings,
			pii_findings,
			injection_findings,
			output_findings
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
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
		secret,
		pii,
		injection,
		output,
	)
	if err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

func (s *PostgresStore) List(
	ctx context.Context,
	limit int,
) ([]Record, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(
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
			secret_findings,
			pii_findings,
			injection_findings,
			output_findings
		FROM audit_events
		ORDER BY id DESC
		LIMIT $1
		`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	results := make([]Record, 0)

	for rows.Next() {
		var item Record
		var timestamp time.Time
		var secret []byte
		var pii []byte
		var injection []byte
		var output []byte

		if err := rows.Scan(
			&item.ID,
			&item.RequestID,
			&timestamp,
			&item.Provider,
			&item.Model,
			&item.Decision,
			&item.RiskScore,
			&item.Severity,
			&item.LatencyMS,
			&secret,
			&pii,
			&injection,
			&output,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}

		item.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
		item.SecretFindings = rawJSON(secret)
		item.PIIFindings = rawJSON(pii)
		item.InjectionFindings = rawJSON(injection)
		item.OutputFindings = rawJSON(output)

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}

	return results, nil
}

func (s *PostgresStore) Stats(
	ctx context.Context,
) (Stats, error) {
	var result Stats

	err := s.pool.QueryRow(
		ctx,
		`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE decision = 'ALLOW'),
			COUNT(*) FILTER (
				WHERE decision = 'ALLOW_WITH_REDACTION'
			),
			COUNT(*) FILTER (WHERE decision = 'BLOCK'),
			COALESCE(AVG(latency_ms), 0),
			COALESCE(AVG(risk_score), 0),
			COUNT(*) FILTER (WHERE severity = 'CRITICAL'),
			COUNT(*) FILTER (WHERE severity = 'HIGH'),
			COUNT(*) FILTER (WHERE severity = 'MEDIUM')
		FROM audit_events
		`,
	).Scan(
		&result.TotalRequests,
		&result.AllowedRequests,
		&result.RedactedRequests,
		&result.BlockedRequests,
		&result.AverageLatencyMS,
		&result.AverageRiskScore,
		&result.CriticalEvents,
		&result.HighSeverityEvents,
		&result.MediumSeverityEvents,
	)
	if err != nil {
		return Stats{}, fmt.Errorf(
			"query audit stats: %w",
			err,
		)
	}

	return result, nil
}

func (s *PostgresStore) WriteToolEvent(
	ctx context.Context,
	event ToolEvent,
) error {
	details, err := marshalJSON(event.Details)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		ctx,
		`
		INSERT INTO tool_events (
			timestamp,
			approval_id,
			event_type,
			tool,
			risk,
			status,
			details
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		event.Timestamp.UTC(),
		event.ApprovalID,
		event.EventType,
		event.Tool,
		event.Risk,
		event.Status,
		details,
	)
	if err != nil {
		return fmt.Errorf("write tool event: %w", err)
	}

	return nil
}

func (s *PostgresStore) ListToolEvents(
	ctx context.Context,
	approvalID int64,
) ([]ToolEventRecord, error) {
	rows, err := s.pool.Query(
		ctx,
		`
		SELECT
			id,
			timestamp,
			approval_id,
			event_type,
			tool,
			risk,
			status,
			details
		FROM tool_events
		WHERE approval_id = $1
		ORDER BY id ASC
		`,
		approvalID,
	)
	if err != nil {
		return nil, fmt.Errorf("query tool events: %w", err)
	}
	defer rows.Close()

	results := make([]ToolEventRecord, 0)

	for rows.Next() {
		var item ToolEventRecord
		var timestamp time.Time
		var details []byte

		if err := rows.Scan(
			&item.ID,
			&timestamp,
			&item.ApprovalID,
			&item.EventType,
			&item.Tool,
			&item.Risk,
			&item.Status,
			&details,
		); err != nil {
			return nil, err
		}

		item.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
		item.Details = rawJSON(details)
		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *PostgresStore) WriteAuthEvent(
	ctx context.Context,
	event AuthEvent,
) error {
	actor, err := marshalJSON(event.Actor)
	if err != nil {
		return err
	}

	details, err := marshalJSON(event.Details)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		ctx,
		`
		INSERT INTO auth_events (
			timestamp,
			event_type,
			key_id,
			key_name,
			actor,
			details
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		event.Timestamp.UTC(),
		event.EventType,
		event.KeyID,
		event.KeyName,
		actor,
		details,
	)
	if err != nil {
		return fmt.Errorf("write auth event: %w", err)
	}

	return nil
}

func (s *PostgresStore) ListAuthEvents(
	ctx context.Context,
	limit int,
) ([]AuthEventRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(
		ctx,
		`
		SELECT
			id,
			timestamp,
			event_type,
			key_id,
			key_name,
			actor,
			details
		FROM auth_events
		ORDER BY id DESC
		LIMIT $1
		`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AuthEventRecord, 0)

	for rows.Next() {
		var item AuthEventRecord
		var timestamp time.Time
		var actor []byte
		var details []byte

		if err := rows.Scan(
			&item.ID,
			&timestamp,
			&item.EventType,
			&item.KeyID,
			&item.KeyName,
			&actor,
			&details,
		); err != nil {
			return nil, err
		}

		item.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
		item.Actor = rawJSON(actor)
		item.Details = rawJSON(details)

		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *PostgresStore) WriteRAGEvent(
	ctx context.Context,
	event RAGEvent,
) error {
	trace, err := json.Marshal(event.Trace)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		ctx,
		`
		INSERT INTO rag_events (
			request_id,
			timestamp,
			user_id,
			role,
			team,
			trace
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		event.RequestID,
		event.Timestamp.UTC(),
		event.UserID,
		event.Role,
		event.Team,
		trace,
	)
	if err != nil {
		return fmt.Errorf("write RAG event: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetRAGEvents(
	ctx context.Context,
	requestID string,
) ([]RAGEventRecord, error) {
	rows, err := s.pool.Query(
		ctx,
		`
		SELECT
			id,
			request_id,
			timestamp,
			user_id,
			role,
			team,
			trace
		FROM rag_events
		WHERE request_id = $1
		ORDER BY id ASC
		`,
		requestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]RAGEventRecord, 0)

	for rows.Next() {
		var item RAGEventRecord
		var timestamp time.Time
		var trace []byte

		if err := rows.Scan(
			&item.ID,
			&item.RequestID,
			&timestamp,
			&item.UserID,
			&item.Role,
			&item.Team,
			&trace,
		); err != nil {
			return nil, err
		}

		item.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
		item.Trace = rawJSON(trace)
		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *PostgresStore) WriteAbuseEvent(
	ctx context.Context,
	event AbuseEvent,
) error {
	details, err := marshalJSON(event.Details)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		ctx,
		`
		INSERT INTO abuse_events (
			timestamp,
			event_type,
			key_id,
			key_name,
			user_id,
			role,
			team,
			path,
			details
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
		`,
		event.Timestamp.UTC(),
		event.EventType,
		event.KeyID,
		event.KeyName,
		event.UserID,
		event.Role,
		event.Team,
		event.Path,
		details,
	)
	if err != nil {
		return fmt.Errorf("write abuse event: %w", err)
	}

	return nil
}

func (s *PostgresStore) ListAbuseEvents(
	ctx context.Context,
	limit int,
) ([]AbuseEventRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.pool.Query(
		ctx,
		`
		SELECT
			id,
			timestamp,
			event_type,
			key_id,
			key_name,
			user_id,
			role,
			team,
			path,
			details
		FROM abuse_events
		ORDER BY id DESC
		LIMIT $1
		`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AbuseEventRecord, 0)

	for rows.Next() {
		var item AbuseEventRecord
		var timestamp time.Time
		var details []byte

		if err := rows.Scan(
			&item.ID,
			&timestamp,
			&item.EventType,
			&item.KeyID,
			&item.KeyName,
			&item.UserID,
			&item.Role,
			&item.Team,
			&item.Path,
			&details,
		); err != nil {
			return nil, err
		}

		item.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
		item.Details = rawJSON(details)
		results = append(results, item)
	}

	return results, rows.Err()
}
