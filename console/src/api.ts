import type {
  APIKeyRecord,
  Approval,
  AuditEvent,
  AuditStats,
  ChatResult,
  CreatedAPIKey,
  EvalResults,
  RAGContextResult,
  RAGDocument,
  RAGInspection,
  SecurityPosture,
  ToolEvaluation,
  ToolEvent,
} from './types'

const API_BASE =
  import.meta.env.VITE_API_BASE ?? ''

const API_KEY_STORAGE = 'sentrymesh_api_key'

export function setAPIKey(key: string) {
  sessionStorage.setItem(
    API_KEY_STORAGE,
    key.trim(),
  )
}

export function clearAPIKey() {
  sessionStorage.removeItem(API_KEY_STORAGE)
}

export function hasAPIKey(): boolean {
  return Boolean(
    sessionStorage.getItem(API_KEY_STORAGE),
  )
}

function authHeaders(
  extra: Record<string, string> = {},
): Record<string, string> {
  const key =
    sessionStorage.getItem(API_KEY_STORAGE)

  return {
    ...extra,
    ...(key
      ? {
          Authorization: `Bearer ${key}`,
        }
      : {}),
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch(
    `${API_BASE}${path}`,
    {
      ...options,
      headers: authHeaders({
        ...(options.body
          ? {
              'Content-Type':
                'application/json',
            }
          : {}),
        ...(options.headers as
          Record<string, string> | undefined),
      }),
    },
  )

  if (!response.ok) {
    let message =
      `request failed: ${response.status}`

    try {
      const body =
        await response.json()

      if (body?.error) {
        message = body.error
      }
    } catch {
      // keep fallback
    }

    throw new Error(message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export function fetchStats() {
  return request<AuditStats>(
    '/v1/audit/stats',
  )
}

export function fetchEvents() {
  return request<AuditEvent[]>(
    '/v1/audit/events?limit=100',
  )
}

export function fetchApprovals() {
  return request<Approval[]>(
    '/v1/approvals',
  )
}

export function approveRequest(
  id: number,
) {
  return request<void>(
    `/v1/approvals/${id}/approve`,
    {
      method: 'POST',
    },
  )
}

export function rejectRequest(
  id: number,
) {
  return request<void>(
    `/v1/approvals/${id}/reject`,
    {
      method: 'POST',
    },
  )
}

export function executeRequest(
  id: number,
) {
  return request<void>(
    `/v1/approvals/${id}/execute`,
    {
      method: 'POST',
    },
  )
}

export function fetchToolEvents(
  approvalId: number,
) {
  return request<ToolEvent[]>(
    `/v1/approvals/${approvalId}/events`,
  )
}

export function fetchEvalResults() {
  return request<EvalResults>(
    '/v1/evals/latest',
  )
}

export function fetchSecurityPosture() {
  return request<SecurityPosture>(
    '/v1/security/posture',
  )
}

export function submitPrompt(
  prompt: string,
) {
  return request<ChatResult>(
    '/v1/chat/completions',
    {
      method: 'POST',
      body: JSON.stringify({
        model: 'test',
        messages: [
          {
            role: 'user',
            content: prompt,
          },
        ],
      }),
    },
  )
}

export function evaluateTool(
  name: string,
  args: Record<string, unknown>,
) {
  return request<ToolEvaluation>(
    '/v1/tools/evaluate',
    {
      method: 'POST',
      body: JSON.stringify({
        name,
        arguments: args,
      }),
    },
  )
}

export function inspectRAG(
  documents: RAGDocument[],
) {
  return request<RAGInspection[]>(
    '/v1/rag/inspect',
    {
      method: 'POST',
      body: JSON.stringify({
        documents,
      }),
    },
  )
}

export function buildRAGContext(
  requestId: string,
  documents: RAGDocument[],
) {
  return request<RAGContextResult>(
    '/v1/rag/context',
    {
      method: 'POST',
      body: JSON.stringify({
        request_id: requestId,
        documents,
      }),
    },
  )
}

export function fetchAPIKeys() {
  return request<APIKeyRecord[]>(
    '/v1/keys',
  )
}

export function createAPIKey(input: {
  name: string
  user_id: string
  role: string
  team: string
  scopes: string[]
  expires_in_hours: number
}) {
  return request<CreatedAPIKey>(
    '/v1/keys',
    {
      method: 'POST',
      body: JSON.stringify(input),
    },
  )
}

export function revokeAPIKey(
  id: number,
) {
  return request<{ id: number; revoked: boolean }>(
    `/v1/keys/${id}/revoke`,
    {
      method: 'POST',
    },
  )
}
