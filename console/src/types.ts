export type Decision =
  | 'ALLOW'
  | 'BLOCK'
  | 'ALLOW_WITH_REDACTION'
  | 'REQUIRE_APPROVAL'
  | string

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

export interface AuditEvent {
  id: number
  timestamp: string
  request_id?: string
  decision?: string
  risk_score?: number
  severity?: string
  provider?: string
  model?: string
  latency_ms?: number
  details?: unknown
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

export interface InjectionFinding {
  type: string
  severity: string
  confidence: number
  action: string
  matched?: string
}

export interface PIIFinding {
  type: string
  severity: string
  action: string
  redacted_as?: string
}

export interface OutputScan {
  safe: boolean
  redacted?: string
}

export interface ChatResult {
  request_id?: string
  decision: Decision
  risk_score: number
  severity: string
  message?: string
  sanitized_prompt?: string
  model_response?: string
  injection_findings?: InjectionFinding[]
  pii_findings?: PIIFinding[]
  output_scan?: OutputScan
}

export interface ToolEvaluation {
  tool: string
  decision: Decision
  reason: string
  risk: number
  approval_id?: number
}

export interface RAGDocument {
  id: string
  source: string
  owner_team: string
  classification: string
  trust_level: string
  content: string
}

export interface RAGInspection {
  document_id: string
  decision: Decision
  sanitized_content?: string
  injection_findings?: InjectionFinding[]
  reason: string
}

export interface RAGTraceEntry {
  document_id: string
  source: string
  trust_level: string
  classification: string
  owner_team: string
  decision: Decision
  included: boolean
  reason: string
}

export interface RAGContextResult {
  context: RAGDocument[]
  trace: {
    request_id: string
    timestamp: string
    entries: RAGTraceEntry[]
  }
}

export interface APIKeyRecord {
  id: number
  name: string
  user_id: string
  role: string
  team: string
  scopes: string
  expires_at?: string
  revoked_at?: string
  created_at: string
}

export interface CreatedAPIKey {
  api_key: string
  name: string
  warning: string
}

export interface SecurityPostureKey {
  key_id: number
  name: string
  user_id: string
  role: string
  team: string
  scopes: string
  abuse_score: number
  status: string
  cooldown_until?: string
  revoked_at?: string
}

export interface SecurityPosture {
  timestamp: string
  summary: {
    total: number
    healthy: number
    elevated: number
    cooldown: number
    revoked: number
  }
  keys: SecurityPostureKey[]
}

export interface EvalMetric {
  total: number
  passed: number
  failed: number
  accuracy: number
  precision?: number
  recall?: number
  false_positives?: number
  false_negatives?: number
  average_latency_ns: number
}

export interface EvalResults {
  timestamp: string
  prompt_injection: EvalMetric
  pii: EvalMetric
  rag: EvalMetric
}
