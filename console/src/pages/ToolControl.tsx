import {
  useState,
} from 'react'

import {
  evaluateTool,
} from '../api'

import type {
  ToolEvaluation,
} from '../types'

const presets: Record<
  string,
  Record<string, unknown>
> = {
  read_customer: {
    fields: ['name'],
  },

  send_email: {
    to: 'customer@example.com',
    subject: 'Production update',
  },

  update_customer: {
    fields: ['company_name'],
  },
}

export default function ToolControl() {
  const [tool, setTool] =
    useState('read_customer')

  const [args, setArgs] =
    useState(
      JSON.stringify(
        presets.read_customer,
        null,
        2,
      ),
    )

  const [result, setResult] =
    useState<ToolEvaluation | null>(
      null,
    )

  const [error, setError] =
    useState('')

  async function run() {
    try {
      setError('')

      setResult(
        await evaluateTool(
          tool,
          JSON.parse(args),
        ),
      )
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'evaluation failed',
      )
    }
  }

  function changeTool(value: string) {
    setTool(value)
    setArgs(
      JSON.stringify(
        presets[value] ?? {},
        null,
        2,
      ),
    )
  }

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>Tool Control</h1>
          <p>
            Evaluate agent actions before
            execution.
          </p>
        </div>
      </div>

      <section className="panel form-grid">
        <label>
          Tool
          <select
            value={tool}
            onChange={(e) =>
              changeTool(
                e.target.value,
              )
            }
          >
            {Object.keys(
              presets,
            ).map((name) => (
              <option key={name}>
                {name}
              </option>
            ))}
          </select>
        </label>

        <label>
          Arguments
          <textarea
            className="json-editor"
            value={args}
            onChange={(e) =>
              setArgs(e.target.value)
            }
          />
        </label>

        <button onClick={run}>
          Evaluate Tool Call
        </button>
      </section>

      {error && (
        <div className="error-box">
          {error}
        </div>
      )}

      {result && (
        <section className="panel">
          <div className="result-header">
            <span
              className={`status ${result.decision
                .toLowerCase()
                .replaceAll(
                  '_',
                  '-',
                )}`}
            >
              {result.decision}
            </span>

            <strong>
              Risk {result.risk}
            </strong>
          </div>

          <p>{result.reason}</p>

          {result.approval_id && (
            <div className="notice">
              Approval #
              {result.approval_id} was
              created. Open the Approvals
              page to review it.
            </div>
          )}
        </section>
      )}
    </div>
  )
}
