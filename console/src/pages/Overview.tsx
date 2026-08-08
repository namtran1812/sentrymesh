import { useEffect, useState } from 'react'

import {
  fetchEvalResults,
  fetchSecurityPosture,
  fetchStats,
} from '../api'

import type {
  AuditStats,
  EvalResults,
  SecurityPosture,
} from '../types'

export default function Overview() {
  const [stats, setStats] =
    useState<AuditStats | null>(null)

  const [evals, setEvals] =
    useState<EvalResults | null>(null)

  const [posture, setPosture] =
    useState<SecurityPosture | null>(null)

  const [error, setError] =
    useState('')

  useEffect(() => {
    Promise.all([
      fetchStats(),
      fetchEvalResults(),
      fetchSecurityPosture(),
    ])
      .then(
        ([
          statsResult,
          evalResult,
          postureResult,
        ]) => {
          setStats(statsResult)
          setEvals(evalResult)
          setPosture(postureResult)
        },
      )
      .catch((err) =>
        setError(err.message),
      )
  }, [])

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Security Overview</h1>
          <p>
            Live gateway health, security
            decisions and evaluation quality.
          </p>
        </div>
      </div>

      {error && (
        <div className="error-box">
          {error}
        </div>
      )}

      <div className="metric-grid">
        <Metric
          label="Requests"
          value={stats?.total_requests ?? 0}
        />

        <Metric
          label="Blocked"
          value={stats?.blocked_requests ?? 0}
        />

        <Metric
          label="Redacted"
          value={
            stats?.redacted_requests ?? 0
          }
        />

        <Metric
          label="Avg Risk"
          value={(
            stats?.average_risk_score ?? 0
          ).toFixed(1)}
        />
      </div>

      <section className="panel">
        <h2>Security Evaluations</h2>

        <div className="metric-grid">
          <Metric
            label="Prompt Injection"
            value={
              evals
                ? `${evals.prompt_injection.passed}/${evals.prompt_injection.total}`
                : '—'
            }
          />

          <Metric
            label="PII"
            value={
              evals
                ? `${evals.pii.passed}/${evals.pii.total}`
                : '—'
            }
          />

          <Metric
            label="RAG"
            value={
              evals
                ? `${evals.rag.passed}/${evals.rag.total}`
                : '—'
            }
          />

          <Metric
            label="Injection Recall"
            value={
              evals?.prompt_injection
                .recall !== undefined
                ? `${(
                    evals
                      .prompt_injection
                      .recall * 100
                  ).toFixed(0)}%`
                : '—'
            }
          />
        </div>
      </section>

      <section className="panel">
        <h2>Credential Posture</h2>

        {posture && (
          <div className="metric-grid">
            <Metric
              label="Healthy"
              value={
                posture.summary.healthy
              }
            />

            <Metric
              label="Elevated"
              value={
                posture.summary.elevated
              }
            />

            <Metric
              label="Cooldown"
              value={
                posture.summary.cooldown
              }
            />

            <Metric
              label="Revoked"
              value={
                posture.summary.revoked
              }
            />
          </div>
        )}
      </section>
    </div>
  )
}

function Metric({
  label,
  value,
}: {
  label: string
  value: string | number
}) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}
