import {
  useEffect,
  useState,
} from 'react'

import {
  fetchEvalResults,
} from '../api'

import type {
  EvalMetric,
  EvalResults,
} from '../types'

export default function Evaluations() {
  const [data, setData] =
    useState<EvalResults | null>(null)

  useEffect(() => {
    fetchEvalResults().then(setData)
  }, [])

  if (!data) {
    return (
      <div className="page">
        Loading evaluations...
      </div>
    )
  }

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Security Evaluations</h1>
          <p>
            Current adversarial regression
            suite results.
          </p>
        </div>
      </div>

      <div className="eval-grid">
        <EvalCard
          name="Prompt Injection"
          metric={
            data.prompt_injection
          }
        />

        <EvalCard
          name="PII Redaction"
          metric={data.pii}
        />

        <EvalCard
          name="RAG Security"
          metric={data.rag}
        />
      </div>

      <section className="panel">
        <h2>Evaluation Snapshot</h2>

        <pre>
          {JSON.stringify(
            data,
            null,
            2,
          )}
        </pre>
      </section>
    </div>
  )
}

function EvalCard({
  name,
  metric,
}: {
  name: string
  metric: EvalMetric
}) {
  return (
    <div className="metric-card eval-card">
      <span>{name}</span>

      <strong>
        {(
          metric.accuracy * 100
        ).toFixed(0)}
        %
      </strong>

      <small>
        {metric.passed}/{metric.total}
        {' passed'}
      </small>
    </div>
  )
}
