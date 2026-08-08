import {
  useState,
} from 'react'

import './App.css'

import APIKeys from './pages/APIKeys'
import Approvals from './pages/Approvals'
import AuditExplorer from './pages/AuditExplorer'
import Evaluations from './pages/Evaluations'
import Overview from './pages/Overview'
import Playground from './pages/Playground'
import Posture from './pages/Posture'
import RAGSecurity from './pages/RAGSecurity'
import ToolControl from './pages/ToolControl'

type Page =
  | 'overview'
  | 'playground'
  | 'rag'
  | 'tools'
  | 'approvals'
  | 'keys'
  | 'posture'
  | 'audit'
  | 'evals'

const navigation: Array<{
  id: Page
  label: string
}> = [
  {
    id: 'overview',
    label: 'Overview',
  },
  {
    id: 'playground',
    label: 'Playground',
  },
  {
    id: 'rag',
    label: 'RAG Security',
  },
  {
    id: 'tools',
    label: 'Tool Control',
  },
  {
    id: 'approvals',
    label: 'Approvals',
  },
  {
    id: 'keys',
    label: 'API Keys',
  },
  {
    id: 'posture',
    label: 'Security Posture',
  },
  {
    id: 'audit',
    label: 'Audit Explorer',
  },
  {
    id: 'evals',
    label: 'Evaluations',
  },
]

export default function App() {
  const [page, setPage] =
    useState<Page>('overview')

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            S
          </div>

          <div>
            <strong>SentryMesh</strong>
            <small>
              Security Console
            </small>
          </div>
        </div>

        <nav>
          {navigation.map((item) => (
            <button
              key={item.id}
              className={
                page === item.id
                  ? 'nav-item active'
                  : 'nav-item'
              }
              onClick={() =>
                setPage(item.id)
              }
            >
              {item.label}
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">
          <span className="live-dot" />
          Production Gateway
        </div>
      </aside>

      <main className="workspace">
        {page === 'overview' && (
          <Overview />
        )}

        {page === 'playground' && (
          <Playground />
        )}

        {page === 'rag' && (
          <RAGSecurity />
        )}

        {page === 'tools' && (
          <ToolControl />
        )}

        {page === 'approvals' && (
          <Approvals />
        )}

        {page === 'keys' && (
          <APIKeys />
        )}

        {page === 'posture' && (
          <Posture />
        )}

        {page === 'audit' && (
          <AuditExplorer />
        )}

        {page === 'evals' && (
          <Evaluations />
        )}
      </main>
    </div>
  )
}
