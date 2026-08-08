import {
  useEffect,
  useMemo,
  useState,
} from 'react'

import {
  fetchEvents,
} from '../api'

import type {
  AuditEvent,
} from '../types'

export default function AuditExplorer() {
  const [events, setEvents] =
    useState<AuditEvent[]>([])

  const [filter, setFilter] =
    useState('')

  const filtered = useMemo(
    () =>
      events.filter((event) =>
        JSON.stringify(event)
          .toLowerCase()
          .includes(
            filter.toLowerCase(),
          ),
      ),
    [events, filter],
  )

  async function refresh() {
    setEvents(await fetchEvents())
  }

  useEffect(() => {
    refresh()
  }, [])

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Audit Explorer</h1>
          <p>
            Inspect security decisions and
            request history.
          </p>
        </div>

        <button
          className="secondary"
          onClick={refresh}
        >
          Refresh
        </button>
      </div>

      <section className="panel">
        <input
          placeholder="Filter events..."
          value={filter}
          onChange={(e) =>
            setFilter(e.target.value)
          }
        />
      </section>

      <section className="panel">
        {filtered.map(
          (event, index) => (
            <details
              className="audit-event"
              key={event.id ?? index}
            >
              <summary>
                <span>
                  {event.decision ??
                    'EVENT'}
                </span>

                <span>
                  {event.severity ?? '—'}
                </span>

                <span>
                  {event.timestamp
                    ? new Date(
                        event.timestamp,
                      ).toLocaleString()
                    : ''}
                </span>
              </summary>

              <pre>
                {JSON.stringify(
                  event,
                  null,
                  2,
                )}
              </pre>
            </details>
          ),
        )}

        {!filtered.length && (
          <div className="empty">
            No audit events.
          </div>
        )}
      </section>
    </div>
  )
}
