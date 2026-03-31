package skills

import (
	"context"
	"encoding/json"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"ab", 1, "a..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestSuspiciousParentChild_Coverage(t *testing.T) {
	// All expected parent processes should be in the map
	expectedParents := []string{"evince", "okular", "chrome", "firefox", "libreoffice"}
	for _, parent := range expectedParents {
		children, ok := SuspiciousParentChild[parent]
		if !ok {
			t.Errorf("expected parent %q in SuspiciousParentChild", parent)
			continue
		}
		if len(children) == 0 {
			t.Errorf("SuspiciousParentChild[%q] has no children", parent)
		}
		// Every parent should flag "bash" and "sh" as suspicious
		hasBash := false
		hasSh := false
		for _, c := range children {
			if c == "bash" {
				hasBash = true
			}
			if c == "sh" {
				hasSh = true
			}
		}
		if !hasBash {
			t.Errorf("SuspiciousParentChild[%q] should include 'bash'", parent)
		}
		if !hasSh {
			t.Errorf("SuspiciousParentChild[%q] should include 'sh'", parent)
		}
	}
}

func TestSuspiciousCommandPatterns_NotEmpty(t *testing.T) {
	if len(SuspiciousCommandPatterns) == 0 {
		t.Fatal("SuspiciousCommandPatterns should not be empty")
	}
	// Should include at least these critical patterns
	required := []string{"nmap", "/etc/passwd", "iptables -F"}
	for _, r := range required {
		found := false
		for _, p := range SuspiciousCommandPatterns {
			if p == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SuspiciousCommandPatterns should include %q", r)
		}
	}
}

func TestLateralMovementDetectorTool_Metadata(t *testing.T) {
	tool := &LateralMovementDetectorTool{}
	if tool.Name() != "detect_lateral_movement" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "detect_lateral_movement")
	}
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
	params := tool.Parameters()
	if params["type"] != "object" {
		t.Errorf("Parameters type should be 'object', got %v", params["type"])
	}
}

func TestLateralMovementDetectorTool_Execute(t *testing.T) {
	// Integration test: runs lateral movement detection on this machine
	tool := &LateralMovementDetectorTool{}
	ctx := context.Background()

	result := tool.Execute(ctx, map[string]any{})
	if result.IsError {
		t.Fatalf("Execute returned error: %s", result.ContentForLLM())
	}

	// Should return valid JSON with the expected structure
	var lmResult LateralMovementResult
	if err := json.Unmarshal([]byte(result.ContentForLLM()), &lmResult); err != nil {
		t.Fatalf("Execute result is not valid JSON: %v", err)
	}

	// Should have scanned at least 1 process
	if lmResult.Scanned < 1 {
		t.Errorf("expected at least 1 process scanned, got %d", lmResult.Scanned)
	}

	t.Logf("Scanned %d processes, found %d suspicious", lmResult.Scanned, len(lmResult.Suspicious))
}
