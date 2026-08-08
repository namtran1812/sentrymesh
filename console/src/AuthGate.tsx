import { useState } from 'react'

import {
  clearAPIKey,
  hasAPIKey,
  setAPIKey,
} from './api'

interface AuthGateProps {
  children: React.ReactNode
}

export default function AuthGate({
  children,
}: AuthGateProps) {
  const [authenticated, setAuthenticated] =
    useState(hasAPIKey())

  const [key, setKey] = useState('')

  if (!authenticated) {
    return (
      <main className="auth-page">
        <section className="auth-card">
          <div className="auth-mark">
            S
          </div>

          <h1>SentryMesh</h1>

          <p>
            Enter an API key to access the
            security console.
          </p>

          <form
            onSubmit={(event) => {
              event.preventDefault()

              const value = key.trim()

              if (!value) {
                return
              }

              setAPIKey(value)
              setAuthenticated(true)
            }}
          >
            <label htmlFor="api-key">
              API Key
            </label>

            <input
              id="api-key"
              type="password"
              value={key}
              autoComplete="off"
              placeholder="sm_..."
              onChange={(event) =>
                setKey(event.target.value)
              }
            />

            <button type="submit">
              Unlock Console
            </button>
          </form>

          <small>
            The key is stored only for this
            browser session.
          </small>
        </section>
      </main>
    )
  }

  return (
    <>
      <button
        className="console-lock"
        onClick={() => {
          clearAPIKey()
          setAuthenticated(false)
          setKey('')
        }}
      >
        Lock
      </button>

      {children}
    </>
  )
}
