import {
  useEffect,
  useState,
} from 'react'

import {
  fetchSecurityPosture,
} from '../api'

import type {
  SecurityPosture,
} from '../types'

export default function Posture() {
  const [data, setData] =
    useState<SecurityPosture | null>(
      null,
    )

  const [error, setError] =
    useState('')

  async function refresh() {
    try {
      setData(
        await fetchSecurityPosture(),
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

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Security Posture</h1>
          <p>
            Credential abuse scores,
            cooldowns and revocations.
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

      {data && (
        <>
          <div className="metric-grid">
            <Metric
              label="Healthy"
              value={
                data.summary.healthy
              }
            />
            <Metric
              label="Elevated"
              value={
                data.summary.elevated
              }
            />
            <Metric
              label="Cooldown"
              value={
                data.summary.cooldown
              }
            />
            <Metric
              label="Revoked"
              value={
                data.summary.revoked
              }
            />
          </div>

          <section className="panel table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Credential</th>
                  <th>User</th>
                  <th>Role</th>
                  <th>Team</th>
                  <th>Score</th>
                  <th>Status</th>
                </tr>
              </thead>

              <tbody>
                {data.keys.map((key) => (
                  <tr key={key.key_id}>
                    <td>{key.name}</td>
                    <td>{key.user_id}</td>
                    <td>{key.role}</td>
                    <td>{key.team}</td>
                    <td>
                      {key.abuse_score}
                    </td>

                    <td>
                      <span
                        className={`status ${key.status.toLowerCase()}`}
                      >
                        {key.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      )}
    </div>
  )
}

function Metric({
  label,
  value,
}: {
  label: string
  value: number
}) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}
