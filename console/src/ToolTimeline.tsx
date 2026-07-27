import type { Approval, ToolEvent } from './types'

interface ToolTimelineProps {
  approval: Approval
  events: ToolEvent[]
  onClose: () => void
}

function ToolTimeline({
  approval,
  events,
  onClose,
}: ToolTimelineProps) {
  return (
    <section className="panel details">
      <div className="panel-header">
        <div>
          <h2>Tool Lifecycle</h2>
          <p className="mono">
            Approval #{approval.id} · {approval.tool}
          </p>
        </div>

        <button onClick={onClose}>
          Close
        </button>
      </div>

      <div className="detail-grid">
        <div>
          <span>Status</span>
          <strong>{approval.status}</strong>
        </div>

        <div>
          <span>Risk</span>
          <strong>{approval.risk}</strong>
        </div>

        <div>
          <span>Tool</span>
          <strong>{approval.tool}</strong>
        </div>

        <div>
          <span>Approval ID</span>
          <strong>#{approval.id}</strong>
        </div>
      </div>

      <div className="timeline">
        {events.length === 0 ? (
          <div className="empty">
            No lifecycle events recorded
          </div>
        ) : (
          events.map((event) => (
            <div
              className="timeline-item"
              key={event.id}
            >
              <div className="timeline-marker" />

              <div className="timeline-content">
                <div className="timeline-top">
                  <strong>{event.event_type}</strong>

                  <span>
                    {new Date(
                      event.timestamp,
                    ).toLocaleTimeString()}
                  </span>
                </div>

                <div className="timeline-meta">
                  Status: {event.status} · Risk:{' '}
                  {event.risk}
                </div>

                <pre>
                  {JSON.stringify(
                    event.details,
                    null,
                    2,
                  )}
                </pre>
              </div>
            </div>
          ))
        )}
      </div>
    </section>
  )
}

export default ToolTimeline
