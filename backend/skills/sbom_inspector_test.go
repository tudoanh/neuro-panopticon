package skills

import (
	"context"
	"testing"
)

func TestSBOMInspector_InvalidPID(t *testing.T) {
	tool := &SBOMInspectorTool{}

	tests := []struct {
		name string
		args map[string]any
	}{
		{"zero", map[string]any{"pid": float64(0)}},
		{"negative", map[string]any{"pid": float64(-1)}},
		{"missing", map[string]any{}},
		{"wrong_type", map[string]any{"pid": "abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.Execute(context.Background(), tt.args)
			if !result.IsError {
				t.Fatal("expected error for invalid PID input")
			}
		})
	}
}

func TestSBOMInspector_Metadata(t *testing.T) {
	tool := &SBOMInspectorTool{}
	if tool.Name() != "inspect_sbom" {
		t.Errorf("expected name 'inspect_sbom', got %q", tool.Name())
	}
	params := tool.Parameters()
	if params["type"] != "object" {
		t.Error("expected parameters type 'object'")
	}
}
