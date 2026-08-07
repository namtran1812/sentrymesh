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

import type { Approval } from './types'

export async function fetchApprovals(): Promise<Approval[]> {
  const response = await fetch(`${API_BASE}/v1/approvals`)

  if (!response.ok) {
    throw new Error(`approvals request failed: ${response.status}`)
  }

  return response.json()
}

export async function approveRequest(id: number): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/approve`,
    { method: 'POST' },
  )

  if (!response.ok) {
    throw new Error(`approve failed: ${response.status}`)
  }
}

export async function rejectRequest(id: number): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/reject`,
    { method: 'POST' },
  )

  if (!response.ok) {
    throw new Error(`reject failed: ${response.status}`)
  }
}

export async function executeRequest(id: number): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/execute`,
    { method: 'POST' },
  )

  if (!response.ok) {
    throw new Error(`execute failed: ${response.status}`)
  }
}

export async function fetchToolEvents(
  approvalId: number,
): Promise<import('./types').ToolEvent[]> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${approvalId}/events`,
  )

  if (!response.ok) {
    throw new Error(
      `tool events request failed: ${response.status}`,
    )
  }

  return response.json()
}

export async function fetchEvalResults(): Promise<
  import('./types').EvalResults
> {
  const response = await fetch(
    `${API_BASE}/v1/evals/latest`,
    {
      headers: {
        Authorization: 'Bearer sm_admin_dev',
      },
    },
  )

  if (!response.ok) {
    throw new Error(
      `eval results request failed: ${response.status}`,
    )
  }

  return response.json()
}
