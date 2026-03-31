package skills

import (
	"context"
	"encoding/json"
	"testing"

	"neuropanopticon/backend/models"
)

func TestProtocolName(t *testing.T) {
	tests := []struct {
		connType uint32
		want     string
	}{
		{1, "tcp"},
		{2, "udp"},
		{0, "unknown(0)"},
		{99, "unknown(99)"},
	}
	for _, tt := range tests {
		got := protocolName(tt.connType)
		if got != tt.want {
			t.Errorf("protocolName(%d) = %q, want %q", tt.connType, got, tt.want)
		}
	}
}

func TestGetProcessName_InvalidPID(t *testing.T) {
	// PID 0 and negative should return empty string
	if name := getProcessName(0); name != "" {
		t.Errorf("getProcessName(0) = %q, want empty", name)
	}
	if name := getProcessName(-1); name != "" {
		t.Errorf("getProcessName(-1) = %q, want empty", name)
	}
}

func TestGetProcessName_CurrentProcess(t *testing.T) {
	// PID 1 (init/systemd) should always exist on Linux
	name := getProcessName(1)
	if name == "" {
		t.Skip("could not read PID 1 name (may need permissions)")
	}
	// Just verify it returned something non-empty
	t.Logf("PID 1 name: %s", name)
}

func TestKnownRiskyPorts_Coverage(t *testing.T) {
	// Verify critical ports are in the map
	criticalPorts := []uint32{22, 23, 80, 443, 3389, 445, 6379, 27017}
	for _, port := range criticalPorts {
		if _, ok := KnownRiskyPorts[port]; !ok {
			t.Errorf("expected port %d to be in KnownRiskyPorts", port)
		}
	}
	// Map should have at least 15 entries
	if len(KnownRiskyPorts) < 15 {
		t.Errorf("KnownRiskyPorts has %d entries, expected at least 15", len(KnownRiskyPorts))
	}
}

func TestFormatScanSummary(t *testing.T) {
	result := &models.NetworkScanResult{
		ListeningPorts: []models.ListeningPort{
			{Protocol: "tcp", Address: "0.0.0.0", Port: 22, PID: 100, ProcessName: "sshd", RiskLevel: "SSH - remote shell access"},
			{Protocol: "tcp", Address: "127.0.0.1", Port: 5432, PID: 200, ProcessName: "postgres"},
		},
		Connections: []models.NetworkConnection{
			{Protocol: "tcp", LocalAddr: "192.168.1.10", LocalPort: 54321, RemoteAddr: "93.184.216.34", RemotePort: 443},
		},
	}

	summary := FormatScanSummary(result)
	if summary == "" {
		t.Fatal("FormatScanSummary returned empty string")
	}
	// Should mention counts
	if !containsStr(summary, "2 listening ports") {
		t.Errorf("summary should mention '2 listening ports', got: %s", summary)
	}
	if !containsStr(summary, "1 established connections") {
		t.Errorf("summary should mention '1 established connections', got: %s", summary)
	}
	// Should mention risk level
	if !containsStr(summary, "SSH") {
		t.Errorf("summary should contain risk level 'SSH', got: %s", summary)
	}
}

func TestFormatScanSummary_Empty(t *testing.T) {
	result := &models.NetworkScanResult{}
	summary := FormatScanSummary(result)
	if !containsStr(summary, "0 listening ports") {
		t.Errorf("empty scan should show '0 listening ports', got: %s", summary)
	}
}

func TestScanNetworkTool_Metadata(t *testing.T) {
	tool := &ScanNetworkTool{}
	if tool.Name() != "scan_network" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "scan_network")
	}
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
	params := tool.Parameters()
	if params["type"] != "object" {
		t.Errorf("Parameters type should be 'object', got %v", params["type"])
	}
}

func TestScanNetworkTool_Execute(t *testing.T) {
	// Integration test: runs a real network scan on this machine
	tool := &ScanNetworkTool{}
	ctx := context.Background()

	result := tool.Execute(ctx, map[string]any{})
	if result.IsError {
		t.Fatalf("Execute returned error: %s", result.ContentForLLM())
	}

	// Should return valid JSON
	var scanResult models.NetworkScanResult
	if err := json.Unmarshal([]byte(result.ContentForLLM()), &scanResult); err != nil {
		t.Fatalf("Execute result is not valid JSON: %v", err)
	}

	// On any real machine there should be at least one listening port
	t.Logf("Found %d listening ports, %d connections", len(scanResult.ListeningPorts), len(scanResult.Connections))
}

func TestScanNetworkTool_ExcludeEstablished(t *testing.T) {
	tool := &ScanNetworkTool{}
	ctx := context.Background()

	result := tool.Execute(ctx, map[string]any{"include_established": false})
	if result.IsError {
		t.Fatalf("Execute returned error: %s", result.ContentForLLM())
	}

	var scanResult models.NetworkScanResult
	if err := json.Unmarshal([]byte(result.ContentForLLM()), &scanResult); err != nil {
		t.Fatalf("Execute result is not valid JSON: %v", err)
	}

	// With include_established=false, connections slice should be empty
	if len(scanResult.Connections) > 0 {
		t.Errorf("expected 0 connections with include_established=false, got %d", len(scanResult.Connections))
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
