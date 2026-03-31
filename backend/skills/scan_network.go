package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/sipeed/picoclaw/pkg/tools"

	"neuropanopticon/backend/models"
)

// KnownRiskyPorts maps port numbers to risk descriptions.
var KnownRiskyPorts = map[uint32]string{
	21:   "FTP - unencrypted file transfer",
	22:   "SSH - remote shell access",
	23:   "Telnet - unencrypted remote access",
	25:   "SMTP - mail relay",
	53:   "DNS - name resolution",
	80:   "HTTP - unencrypted web server",
	135:  "MSRPC - Windows RPC",
	139:  "NetBIOS - Windows file sharing",
	443:  "HTTPS - web server",
	445:  "SMB - Windows file sharing",
	1433: "MSSQL - database",
	1434: "MSSQL Browser",
	3306: "MySQL - database",
	3389: "RDP - Remote Desktop",
	5432: "PostgreSQL - database",
	5900: "VNC - remote desktop",
	6379: "Redis - in-memory store",
	8080: "HTTP Alt - web server",
	8443: "HTTPS Alt - web server",
	9200: "Elasticsearch",
	27017: "MongoDB - database",
}

type ScanNetworkTool struct{}

func (t *ScanNetworkTool) Name() string        { return "scan_network" }
func (t *ScanNetworkTool) Description() string {
	return "Scans all active listening ports and established network connections on this machine. " +
		"Returns listening ports with process info and risk levels, plus active connections. " +
		"Use this to detect exposed services, suspicious connections, or dangerous open ports."
}

func (t *ScanNetworkTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"include_established": map[string]any{
				"type":        "boolean",
				"description": "Whether to include established connections (not just listening ports). Default: true.",
			},
		},
	}
}

func (t *ScanNetworkTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	includeEstablished := true
	if val, ok := args["include_established"].(bool); ok {
		includeEstablished = val
	}

	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to scan network connections: %v", err))
	}

	result := models.NetworkScanResult{}

	// Collect listening ports
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			port := models.ListeningPort{
				Protocol: protocolName(conn.Type),
				Address:  conn.Laddr.IP,
				Port:     conn.Laddr.Port,
				PID:      conn.Pid,
			}
			port.ProcessName = getProcessName(conn.Pid)
			if risk, known := KnownRiskyPorts[conn.Laddr.Port]; known {
				port.RiskLevel = risk
			}
			result.ListeningPorts = append(result.ListeningPorts, port)
		}
	}

	// Collect established connections
	if includeEstablished {
		for _, conn := range connections {
			if conn.Status == "ESTABLISHED" {
				nc := models.NetworkConnection{
					Protocol:   protocolName(conn.Type),
					LocalAddr:  conn.Laddr.IP,
					LocalPort:  conn.Laddr.Port,
					RemoteAddr: conn.Raddr.IP,
					RemotePort: conn.Raddr.Port,
					Status:     conn.Status,
					PID:        conn.Pid,
				}
				nc.ProcessName = getProcessName(conn.Pid)
				result.Connections = append(result.Connections, nc)
			}
		}
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return tools.NewToolResult(string(data))
}

func protocolName(connType uint32) string {
	switch connType {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	default:
		return fmt.Sprintf("unknown(%d)", connType)
	}
}

func getProcessName(pid int32) string {
	if pid <= 0 {
		return ""
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := p.Name()
	if err != nil {
		return ""
	}
	return name
}

// FormatScanSummary produces a human-readable summary of a network scan.
func FormatScanSummary(result *models.NetworkScanResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d listening ports and %d established connections.\n\n",
		len(result.ListeningPorts), len(result.Connections)))

	if len(result.ListeningPorts) > 0 {
		b.WriteString("## Listening Ports\n")
		for _, lp := range result.ListeningPorts {
			risk := ""
			if lp.RiskLevel != "" {
				risk = fmt.Sprintf(" [%s]", lp.RiskLevel)
			}
			b.WriteString(fmt.Sprintf("- %s:%d (%s) PID=%d%s\n",
				lp.Address, lp.Port, lp.ProcessName, lp.PID, risk))
		}
	}

	return b.String()
}
