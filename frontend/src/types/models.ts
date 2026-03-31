export type Severity = "critical" | "high" | "medium" | "low" | "info";

export interface Finding {
  id: string;
  title: string;
  description: string;
  severity: Severity;
  category: string;
  source: string;
  remediation?: string;
  timestamp: string;
}

export interface SecurityScore {
  score: number;
  max_score: number;
  findings: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  last_updated: string;
}

export interface SystemStatus {
  cpu_percent: number;
  memory_used_gb: number;
  memory_total_gb: number;
  network_in_bytes: number;
  network_out_bytes: number;
  security_score: SecurityScore;
  agent_ready: boolean;
  llm_backend: string;
}

export interface ToolCall {
  id: string;
  name: string;
  args: string;
  result?: string;
  status: string;
  duration_ms?: number;
}

export interface ChatResponse {
  content: string;
  tool_calls?: ToolCall[];
}

export interface LLMConfig {
  backend: string;
  ollama_url: string;
  ollama_model: string;
  cloud_api_key?: string;
  cloud_url?: string;
  cloud_model?: string;
  max_tokens: number;
  temperature: number;
}

export interface SecurityConfig {
  allow_remediation: boolean;
  allowed_remediations: string[];
  scan_interval_seconds: number;
}

export interface AppConfig {
  llm: LLMConfig;
  security: SecurityConfig;
}
