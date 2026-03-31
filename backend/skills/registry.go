package skills

import (
	"neuropanopticon/backend/config"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// RegisterAll registers all NeuroPanopticon security skills into a picoclaw ToolRegistry.
func RegisterAll(registry *tools.ToolRegistry, cfg *config.AppConfig) {
	registry.Register(&ScanNetworkTool{})
	registry.Register(&LateralMovementDetectorTool{})
	registry.Register(&BlastRadiusAnalyzerTool{})
	registry.Register(&SBOMInspectorTool{})
	registry.Register(&RemediateTool{
		Enabled:             cfg.Security.AllowRemediation,
		AllowedRemediations: cfg.Security.AllowedRemediations,
	})
}
