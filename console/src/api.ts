import type {
  Approval,
  AuditEvent,
  AuditStats,
  EvalResults,
  ToolEvent,
} from './types'

const API_BASE =
  import.meta.env.VITE_API_BASE ?? ''

const API_KEY_STORAGE =
  'sentrymesh_api_key'

export function setAPIKey(
  key: string,
) {
  sessionStorage.setItem(
    API_KEY_STORAGE,
    key.trim(),
  )
}

export function clearAPIKey() {
  sessionStorage.removeItem(
    API_KEY_STORAGE,
  )
}

export function hasAPIKey(): boolean {
  return Boolean(
    sessionStorage.getItem(
      API_KEY_STORAGE,
    ),
  )
}

function authHeaders(
  extra: Record<string, string> = {},
): Record<string, string> {
  const key = sessionStorage.getItem(
    API_KEY_STORAGE,
  )

  return {
    ...extra,
    ...(key
      ? {
          Authorization: `Bearer ${key}`,
        }
      : {}),
  }
}

async function requireOK(
  response: Response,
  description: string,
) {
  if (!response.ok) {
    throw new Error(
      `${description}: ${response.status}`,
    )
  }
}

export async function fetchStats():
Promise<AuditStats> {
  const response = await fetch(
    `${API_BASE}/v1/audit/stats`,
    {
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'stats request failed',
  )

  return response.json()
}

export async function fetchEvents():
Promise<AuditEvent[]> {
  const response = await fetch(
    `${API_BASE}/v1/audit/events?limit=50`,
    {
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'events request failed',
  )

  return response.json()
}

export async function fetchApprovals():
Promise<Approval[]> {
  const response = await fetch(
    `${API_BASE}/v1/approvals`,
    {
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'approvals request failed',
  )

  return response.json()
}

export async function approveRequest(
  id: number,
): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/approve`,
    {
      method: 'POST',
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'approve failed',
  )
}

export async function rejectRequest(
  id: number,
): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/reject`,
    {
      method: 'POST',
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'reject failed',
  )
}

export async function executeRequest(
  id: number,
): Promise<void> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${id}/execute`,
    {
      method: 'POST',
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'execute failed',
  )
}

export async function fetchToolEvents(
  approvalId: number,
): Promise<ToolEvent[]> {
  const response = await fetch(
    `${API_BASE}/v1/approvals/${approvalId}/events`,
    {
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'tool events request failed',
  )

  return response.json()
}

export async function fetchEvalResults():
Promise<EvalResults> {
  const response = await fetch(
    `${API_BASE}/v1/evals/latest`,
    {
      headers: authHeaders(),
    },
  )

  await requireOK(
    response,
    'eval results request failed',
  )

  return response.json()
}
