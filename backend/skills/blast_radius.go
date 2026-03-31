package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/sipeed/picoclaw/pkg/tools"

	"neuropanopticon/backend/models"
)

// VPN interface name patterns across platforms.
var vpnPatterns = []string{
	"tun", "tap", "wg", "utun", "ppp", "vpn",
	"wireguard", "nordlynx", "proton", "mullvad",
}

type BlastRadiusAnalyzerTool struct{}

func (t *BlastRadiusAnalyzerTool) Name() string { return "analyze_blast_radius" }
func (t *BlastRadiusAnalyzerTool) Description() string {
	return "Analyzes the network environment to determine blast radius for threat scoring. " +
		"Checks for VPN connections, public WiFi, exposed ports, and network interfaces. " +
		"Use this to provide context-aware threat scoring."
}

func (t *BlastRadiusAnalyzerTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *BlastRadiusAnalyzerTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	interfaces, err := psnet.InterfacesWithContext(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get network interfaces: %v", err))
	}

	result := models.BlastRadiusInfo{
		RiskMultiplier: 1.0,
	}

	for _, iface := range interfaces {
		if len(iface.Addrs) == 0 {
			continue
		}

		ni := models.NetworkInterface{
			Name: iface.Name,
			IsUp: hasFlag(iface.Flags, "up"),
		}

		for _, addr := range iface.Addrs {
			ni.Addresses = append(ni.Addresses, addr.Addr)
		}

		// Check if this is a VPN interface
		nameLower := strings.ToLower(iface.Name)
		for _, pattern := range vpnPatterns {
			if strings.Contains(nameLower, pattern) {
				ni.IsVPN = true
				result.IsVPNActive = true
				break
			}
		}

		result.Interfaces = append(result.Interfaces, ni)
	}

	// Count exposed ports (listening on 0.0.0.0 or ::)
	connections, err := psnet.ConnectionsWithContext(ctx, "all")
	if err == nil {
		for _, conn := range connections {
			if conn.Status == "LISTEN" {
				if conn.Laddr.IP == "0.0.0.0" || conn.Laddr.IP == "::" || conn.Laddr.IP == "" {
					result.ExposedPorts++
				}
			}
		}
	}

	// Calculate risk multiplier
	if result.IsVPNActive {
		result.RiskMultiplier *= 1.5
	}
	if result.ExposedPorts > 5 {
		result.RiskMultiplier *= 1.3
	}
	if result.ExposedPorts > 10 {
		result.RiskMultiplier *= 1.5
	}

	// Build summary
	var parts []string
	if result.IsVPNActive {
		parts = append(parts, "VPN connection active (corporate network pivot risk)")
	}
	parts = append(parts, fmt.Sprintf("%d ports exposed to all interfaces", result.ExposedPorts))
	parts = append(parts, fmt.Sprintf("%d network interfaces detected", len(result.Interfaces)))
	parts = append(parts, fmt.Sprintf("Risk multiplier: %.1fx", result.RiskMultiplier))
	result.Summary = strings.Join(parts, "; ")

	data, _ := json.MarshalIndent(result, "", "  ")
	return tools.NewToolResult(string(data))
}

func hasFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if strings.EqualFold(f, flag) {
			return true
		}
	}
	return false
}
