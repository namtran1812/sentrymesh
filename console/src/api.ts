import type { AuditEvent, AuditStats } from './types'

const API_BASE = 'http://localhost:8080'

export async function fetchStats(): Promise<AuditStats> {
  const response = await fetch(`${API_BASE}/v1/audit/stats`)

  if (!response.ok) {
    throw new Error(`stats request failed: ${response.status}`)
  }

  return response.json()
}

export async function fetchEvents(): Promise<AuditEvent[]> {
  const response = await fetch(`${API_BASE}/v1/audit/events?limit=50`)

  if (!response.ok) {
    throw new Error(`events request failed: ${response.status}`)
  }

  return response.json()
}
