package audit

import "context"

// Repository defines the persistence contract used by the gateway for
// security audit data. Implementations may use SQLite, PostgreSQL, or
// another durable store without leaking database details into the API
// and middleware layers.
type Repository interface {
	Write(context.Context, Event) error
	List(context.Context, int) ([]Record, error)
	Stats(context.Context) (Stats, error)

	WriteToolEvent(context.Context, ToolEvent) error
	ListToolEvents(context.Context, int64) ([]ToolEventRecord, error)

	WriteAuthEvent(context.Context, AuthEvent) error
	ListAuthEvents(context.Context, int) ([]AuthEventRecord, error)

	WriteRAGEvent(context.Context, RAGEvent) error
	GetRAGEvents(context.Context, string) ([]RAGEventRecord, error)

	WriteAbuseEvent(context.Context, AbuseEvent) error
	ListAbuseEvents(context.Context, int) ([]AbuseEventRecord, error)
}

// Compile-time assertion that the SQLite implementation continues to
// satisfy the repository contract.
var _ Repository = (*Store)(nil)
