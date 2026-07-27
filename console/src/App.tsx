import { useEffect, useState } from 'react'
import { fetchEvents, fetchStats } from './api'
import type { AuditEvent, AuditStats } from './types'
import './App.css'

function StatCard({
  label,
  value,
}: {
  label: string
  value: string | number
}) {
  return (
    <div className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
    </div>
  )
}

function App() {
  const [stats, setStats] = useState<AuditStats | null>(null)
  const [events, setEvents] = useState<AuditEvent[]>([])
  const [selected, setSelected] = useState<AuditEvent | null>(null)
  const [error, setError] = useState('')

  async function loadData() {
    try {
      const [statsData, eventsData] = await Promise.all([
        fetchStats(),
        fetchEvents(),
      ])

      setStats(statsData)
      setEvents(eventsData)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'unknown error')
    }
  }

  useEffect(() => {
    void loadData()

    const timer = window.setInterval(() => {
      void loadData()
    }, 5000)

    return () => window.clearInterval(timer)
  }, [])

  return (
    <main>
      <header>
        <div>
          <h1>SentryMesh</h1>
          <p>AI Security Operations</p>
        </div>

        <button onClick={() => void loadData()}>
          Refresh
        </button>
      </header>

      {error && <div className="error">{error}</div>}

      <section className="stats">
        <StatCard
          label="Total Requests"
          value={stats?.total_requests ?? '-'}
        />
        <StatCard
          label="Allowed"
          value={stats?.allowed_requests ?? '-'}
        />
        <StatCard
          label="Redacted"
          value={stats?.redacted_requests ?? '-'}
        />
        <StatCard
          label="Blocked"
          value={stats?.blocked_requests ?? '-'}
        />
        <StatCard
          label="Average Risk"
          value={stats?.average_risk_score.toFixed(1) ?? '-'}
        />
        <StatCard
          label="Average Latency"
          value={
            stats
              ? `${stats.average_latency_ms.toFixed(0)} ms`
              : '-'
          }
        />
      </section>

      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Security Events</h2>
            <p>Latest gateway decisions and detections</p>
          </div>

          <div className="live">
            <span />
            LIVE
          </div>
        </div>

        <div className="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Request</th>
                <th>Provider</th>
                <th>Model</th>
                <th>Decision</th>
                <th>Risk</th>
                <th>Latency</th>
              </tr>
            </thead>

            <tbody>
              {events.map((event) => (
                <tr
                  key={event.id}
                  onClick={() => setSelected(event)}
                >
                  <td>
                    {new Date(event.timestamp).toLocaleTimeString()}
                  </td>
                  <td className="mono">{event.request_id}</td>
                  <td>{event.provider}</td>
                  <td>{event.model}</td>
                  <td>
                    <span
                      className={`badge ${event.decision.toLowerCase()}`}
                    >
                      {event.decision}
                    </span>
                  </td>
                  <td>{event.risk_score}</td>
                  <td>{event.latency_ms} ms</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {selected && (
        <section className="panel details">
          <div className="panel-header">
            <div>
              <h2>Request Trace</h2>
              <p className="mono">{selected.request_id}</p>
            </div>

            <button onClick={() => setSelected(null)}>
              Close
            </button>
          </div>

          <div className="detail-grid">
            <div>
              <span>Decision</span>
              <strong>{selected.decision}</strong>
            </div>

            <div>
              <span>Severity</span>
              <strong>{selected.severity}</strong>
            </div>

            <div>
              <span>Risk Score</span>
              <strong>{selected.risk_score}</strong>
            </div>

            <div>
              <span>Latency</span>
              <strong>{selected.latency_ms} ms</strong>
            </div>
          </div>

          <h3>PII Findings</h3>
          <pre>
            {JSON.stringify(selected.pii_findings, null, 2)}
          </pre>

          <h3>Injection Findings</h3>
          <pre>
            {JSON.stringify(selected.injection_findings, null, 2)}
          </pre>

          <h3>Secret Findings</h3>
          <pre>
            {JSON.stringify(selected.secret_findings, null, 2)}
          </pre>

          <h3>Output Scan</h3>
          <pre>
            {JSON.stringify(selected.output_findings, null, 2)}
          </pre>
        </section>
      )}
    </main>
  )
}

export default App
