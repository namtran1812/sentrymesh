import {
  useEffect,
  useState,
} from 'react'

import {
  createAPIKey,
  fetchAPIKeys,
  revokeAPIKey,
} from '../api'

import type {
  APIKeyRecord,
  CreatedAPIKey,
} from '../types'

const availableScopes = [
  'tools:evaluate',
  'approvals:write',
  'tools:execute',
  'audit:read',
  'keys:manage',
  'rag:inspect',
  'rag:context',
  'rag:chat',
  'evals:read',
]

export default function APIKeys() {
  const [keys, setKeys] =
    useState<APIKeyRecord[]>([])

  const [created, setCreated] =
    useState<CreatedAPIKey | null>(
      null,
    )

  const [name, setName] =
    useState('demo-client')

  const [userId, setUserId] =
    useState('u_demo_client')

  const [role, setRole] =
    useState('analyst')

  const [team, setTeam] =
    useState('risk')

  const [scopes, setScopes] =
    useState<string[]>([
      'tools:evaluate',
      'rag:context',
    ])

  const [error, setError] =
    useState('')

  async function refresh() {
    try {
      setKeys(await fetchAPIKeys())
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

  async function create() {
    try {
      setError('')

      const result =
        await createAPIKey({
          name,
          user_id: userId,
          role,
          team,
          scopes,
          expires_in_hours: 24,
        })

      setCreated(result)
      await refresh()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'create failed',
      )
    }
  }

  async function revoke(id: number) {
    try {
      await revokeAPIKey(id)
      await refresh()
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'revoke failed',
      )
    }
  }

  return (
    <div className="page">
      <div className="page-title">
        <div>
          <h1>API Keys</h1>
          <p>
            Create scoped credentials and
            revoke compromised access.
          </p>
        </div>
      </div>

      {created && (
        <section className="panel secret-panel">
          <h2>
            New API Key
          </h2>

          <p>
            This value is displayed once.
          </p>

          <code className="secret">
            {created.api_key}
          </code>

          <button
            className="secondary"
            onClick={() =>
              navigator.clipboard.writeText(
                created.api_key,
              )
            }
          >
            Copy
          </button>
        </section>
      )}

      <section className="panel form-grid">
        <h2>Create Credential</h2>

        <label>
          Name
          <input
            value={name}
            onChange={(e) =>
              setName(e.target.value)
            }
          />
        </label>

        <label>
          User ID
          <input
            value={userId}
            onChange={(e) =>
              setUserId(e.target.value)
            }
          />
        </label>

        <label>
          Role
          <select
            value={role}
            onChange={(e) =>
              setRole(e.target.value)
            }
          >
            <option value="analyst">
              analyst
            </option>
            <option value="sales">
              sales
            </option>
            <option value="admin">
              admin
            </option>
          </select>
        </label>

        <label>
          Team
          <input
            value={team}
            onChange={(e) =>
              setTeam(e.target.value)
            }
          />
        </label>

        <div>
          <span className="field-label">
            Scopes
          </span>

          <div className="scope-grid">
            {availableScopes.map(
              (scope) => (
                <label
                  className="checkbox"
                  key={scope}
                >
                  <input
                    type="checkbox"
                    checked={scopes.includes(
                      scope,
                    )}
                    onChange={() =>
                      setScopes((current) =>
                        current.includes(
                          scope,
                        )
                          ? current.filter(
                              (value) =>
                                value !==
                                scope,
                            )
                          : [
                              ...current,
                              scope,
                            ],
                      )
                    }
                  />

                  {scope}
                </label>
              ),
            )}
          </div>
        </div>

        <button onClick={create}>
          Create API Key
        </button>
      </section>

      {error && (
        <div className="error-box">
          {error}
        </div>
      )}

      <section className="panel table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>User</th>
              <th>Role</th>
              <th>Team</th>
              <th>Scopes</th>
              <th>Status</th>
              <th />
            </tr>
          </thead>

          <tbody>
            {keys.map((key) => (
              <tr key={key.id}>
                <td>{key.name}</td>
                <td>{key.user_id}</td>
                <td>{key.role}</td>
                <td>{key.team}</td>
                <td className="mono-small">
                  {key.scopes}
                </td>

                <td>
                  {key.revoked_at
                    ? 'REVOKED'
                    : 'ACTIVE'}
                </td>

                <td>
                  {!key.revoked_at && (
                    <button
                      className="danger"
                      onClick={() =>
                        revoke(key.id)
                      }
                    >
                      Revoke
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  )
}
