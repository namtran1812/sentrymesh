import {
  useEffect,
  useState,
} from 'react'

import {
  approveRequest,
  executeRequest,
  fetchApprovals,
  fetchToolEvents,
  rejectRequest,
} from '../api'

import type {
  Approval,
  ToolEvent,
} from '../types'

export default function Approvals() {
  const [items, setItems] =
    useState<Approval[]>([])

  const [events, setEvents] =
    useState<ToolEvent[]>([])

  const [selected, setSelected] =
    useState<number | null>(null)

  const [error, setError] =
    useState('')

  async function refresh() {
    try {
      setItems(
        await fetchApprovals(),
      )
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'failed',
      )
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  async function lifecycle(id: number) {
    setSelected(id)
    setEvents(
      await fetchToolEvents(id),
    )
  }

  async function action(
    fn: (id: number) =>
      Promise<unknown>,
    id: number,
  ) {
    try {
      await fn(id)
      await refresh()

      if (selected === id) {
        await lifecycle(id)
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'action failed',
      )
    }
  }

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Approvals</h1>
          <p>
            Human review for sensitive
            agent actions.
          </p>
        </div>

        <button
          className="secondary"
          onClick={refresh}
        >
          Refresh
        </button>
      </div>

      {error && (
        <div className="error-box">
          {error}
        </div>
      )}

      <section className="panel table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Tool</th>
              <th>Risk</th>
              <th>Status</th>
              <th>Reason</th>
              <th>Actions</th>
            </tr>
          </thead>

          <tbody>
            {items.map((item) => (
              <tr key={item.id}>
                <td>#{item.id}</td>
                <td>{item.tool}</td>
                <td>{item.risk}</td>

                <td>
                  <span
                    className={`status ${item.status.toLowerCase()}`}
                  >
                    {item.status}
                  </span>
                </td>

                <td>{item.reason}</td>

                <td>
                  <div className="table-actions">
                    {item.status ===
                      'PENDING' && (
                      <>
                        <button
                          onClick={() =>
                            action(
                              approveRequest,
                              item.id,
                            )
                          }
                        >
                          Approve
                        </button>

                        <button
                          className="danger"
                          onClick={() =>
                            action(
                              rejectRequest,
                              item.id,
                            )
                          }
                        >
                          Reject
                        </button>
                      </>
                    )}

                    {item.status ===
                      'APPROVED' && (
                      <button
                        onClick={() =>
                          action(
                            executeRequest,
                            item.id,
                          )
                        }
                      >
                        Execute
                      </button>
                    )}

                    <button
                      className="secondary"
                      onClick={() =>
                        lifecycle(
                          item.id,
                        )
                      }
                    >
                      Timeline
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {!items.length && (
          <div className="empty">
            No active approvals.
          </div>
        )}
      </section>

      {selected !== null && (
        <section className="panel">
          <h2>
            Approval #{selected} Lifecycle
          </h2>

          {events.map((event) => (
            <div
              className="timeline-item"
              key={event.id}
            >
              <strong>
                {event.event_type}
              </strong>

              <span>
                {event.status}
                {' · '}
                {new Date(
                  event.timestamp,
                ).toLocaleString()}
              </span>

              <pre>
                {JSON.stringify(
                  event.details,
                  null,
                  2,
                )}
              </pre>
            </div>
          ))}
        </section>
      )}
    </div>
  )
}
