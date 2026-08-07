import type { EvalResults } from './types'

interface EvalMetricsProps {
  results: EvalResults | null
}

function percent(value: number | undefined) {
  if (value === undefined) {
    return '-'
  }

  return `${(value * 100).toFixed(1)}%`
}

function EvalMetrics({
  results,
}: EvalMetricsProps) {
  if (!results) {
    return (
      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Security Evaluations</h2>
            <p>No evaluation data loaded</p>
          </div>
        </div>
      </section>
    )
  }

  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h2>Security Evaluations</h2>
          <p>
            Latest automated adversarial test results
          </p>
        </div>

        <div className="eval-time">
          {new Date(
            results.timestamp,
          ).toLocaleString()}
        </div>
      </div>

      <div className="eval-grid">
        <div className="eval-card">
          <span>Prompt Injection</span>
          <strong>
            {percent(
              results.prompt_injection.accuracy,
            )}
          </strong>

          <small>
            {results.prompt_injection.passed}/
            {results.prompt_injection.total} passed
          </small>
        </div>

        <div className="eval-card">
          <span>PII Redaction</span>
          <strong>
            {percent(results.pii.accuracy)}
          </strong>

          <small>
            {results.pii.passed}/
            {results.pii.total} passed
          </small>
        </div>

        <div className="eval-card">
          <span>RAG Security</span>
          <strong>
            {percent(results.rag.accuracy)}
          </strong>

          <small>
            {results.rag.passed}/
            {results.rag.total} passed
          </small>
        </div>
      </div>

      <div className="eval-details">
        <div>
          <span>Injection Precision</span>
          <strong>
            {percent(
              results.prompt_injection.precision,
            )}
          </strong>
        </div>

        <div>
          <span>Injection Recall</span>
          <strong>
            {percent(
              results.prompt_injection.recall,
            )}
          </strong>
        </div>

        <div>
          <span>False Positives</span>
          <strong>
            {
              results.prompt_injection
                .false_positives ?? 0
            }
          </strong>
        </div>

        <div>
          <span>False Negatives</span>
          <strong>
            {
              results.prompt_injection
                .false_negatives ?? 0
            }
          </strong>
        </div>
      </div>
    </section>
  )
}

export default EvalMetrics
