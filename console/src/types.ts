export interface AuditStats {
  total_requests: number
  allowed_requests: number
  redacted_requests: number
  blocked_requests: number
  average_latency_ms: number
  average_risk_score: number
  critical_events: number
  high_severity_events: number
  medium_severity_events: number
}

export interface Finding {
  type?: string
  severity?: string
  action?: string
  confidence?: number
  redacted_as?: string
  matched?: string
}

export interface OutputScan {
  safe: boolean
  redacted: string
  pii_findings?: Finding[]
  secret_findings?: Finding[]
}

export interface AuditEvent {
  id: number
  request_id: string
  timestamp: string
  provider: string
  model: string
  decision: string
  risk_score: number
  severity: string
  latency_ms: number
  secret_findings: Finding[] | null
  pii_findings: Finding[]
  injection_findings: Finding[]
  output_findings: OutputScan | null
}

export interface Approval {
  id: number
  created_at: string
  tool: string
  arguments: Record<string, unknown>
  risk: number
  reason: string
  status: string
  executed_at?: string
}

export interface ToolEvent {
  id: number
  timestamp: string
  approval_id: number
  event_type: string
  tool: string
  risk: number
  status: string
  details: unknown
}
