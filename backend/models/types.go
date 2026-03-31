package models

import "time"

// Severity levels for security findings.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Finding represents a single security finding from a scan or analysis.
type Finding struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    Severity  `json:"severity"`
	Category    string    `json:"category"`
	Source      string    `json:"source"`
	Remediation string    `json:"remediation,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// NetworkConnection represents an active network connection.
type NetworkConnection struct {
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint32 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint32 `json:"remote_port"`
	Status      string `json:"status"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name"`
}

// ListeningPort represents a port that is actively listening for connections.
type ListeningPort struct {
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	Port        uint32 `json:"port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name"`
	RiskLevel   string `json:"risk_level,omitempty"`
}

// NetworkScanResult holds the results of a full network scan.
type NetworkScanResult struct {
	ListeningPorts []ListeningPort     `json:"listening_ports"`
	Connections    []NetworkConnection `json:"connections"`
	Timestamp      time.Time           `json:"timestamp"`
}

// ProcessInfo represents a running process with security-relevant metadata.
type ProcessInfo struct {
	PID         int32    `json:"pid"`
	Name        string   `json:"name"`
	Cmdline     string   `json:"cmdline"`
	Username    string   `json:"username"`
	ParentPID   int32    `json:"parent_pid"`
	ParentName  string   `json:"parent_name"`
	CreateTime  int64    `json:"create_time"`
	CPUPercent  float64  `json:"cpu_percent"`
	MemoryMB    float64  `json:"memory_mb"`
	Connections int      `json:"connections"`
	Children    []string `json:"children,omitempty"`
}

// BlastRadiusInfo describes the network environment context for threat scoring.
type BlastRadiusInfo struct {
	Interfaces     []NetworkInterface `json:"interfaces"`
	IsVPNActive    bool               `json:"is_vpn_active"`
	IsPublicWiFi   bool               `json:"is_public_wifi"`
	ExposedPorts   int                `json:"exposed_ports"`
	RiskMultiplier float64            `json:"risk_multiplier"`
	Summary        string             `json:"summary"`
}

// NetworkInterface describes a network interface on the system.
type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	IsUp      bool     `json:"is_up"`
	IsVPN     bool     `json:"is_vpn"`
}

// RemediationAction represents a safe, predefined action the agent can take.
type RemediationAction struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Reversible  bool   `json:"reversible"`
}

// SecurityScore represents the overall system security posture.
type SecurityScore struct {
	Score       int       `json:"score"`
	MaxScore    int       `json:"max_score"`
	Findings    int       `json:"findings"`
	Critical    int       `json:"critical"`
	High        int       `json:"high"`
	Medium      int       `json:"medium"`
	Low         int       `json:"low"`
	LastUpdated time.Time `json:"last_updated"`
}

// ChatMessage represents a message in the agent chat interface.
type ChatMessage struct {
	ID        string     `json:"id"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// ToolCall represents a tool invocation by the agent.
type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Args     string `json:"args"`
	Result   string `json:"result,omitempty"`
	Status   string `json:"status"`
	Duration int64  `json:"duration_ms,omitempty"`
}

// SystemStatus represents the current monitoring status shown on the dashboard.
type SystemStatus struct {
	CPUPercent  float64       `json:"cpu_percent"`
	MemoryUsed float64       `json:"memory_used_gb"`
	MemoryTotal float64      `json:"memory_total_gb"`
	NetworkIn  uint64         `json:"network_in_bytes"`
	NetworkOut uint64         `json:"network_out_bytes"`
	Score      SecurityScore  `json:"security_score"`
	AgentReady bool           `json:"agent_ready"`
	LLMBackend string         `json:"llm_backend"`
}

// LLMProvider identifies which LLM backend is in use.
type LLMBackend string

const (
	LLMBackendOllama LLMBackend = "ollama"
	LLMBackendCloud  LLMBackend = "cloud"
)
