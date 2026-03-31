package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/sipeed/picoclaw/pkg/tools"
)

type SBOMEntry struct {
	Library string `json:"library"`
	Path    string `json:"path"`
	Type    string `json:"type"`
}

type SBOMResult struct {
	ProcessName string      `json:"process_name"`
	PID         int32       `json:"pid"`
	Executable  string      `json:"executable"`
	Libraries   []SBOMEntry `json:"libraries"`
	Error       string      `json:"error,omitempty"`
}

type SBOMInspectorTool struct{}

func (t *SBOMInspectorTool) Name() string { return "inspect_sbom" }
func (t *SBOMInspectorTool) Description() string {
	return "Analyzes a running process to extract its Software Bill of Materials (shared libraries and dependencies). " +
		"Provide a process ID to inspect. This helps detect supply-chain attacks by revealing what libraries a process loads."
}

func (t *SBOMInspectorTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pid": map[string]any{
				"type":        "number",
				"description": "The process ID to inspect",
			},
		},
		"required": []string{"pid"},
	}
}

func (t *SBOMInspectorTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	pidFloat, ok := args["pid"].(float64)
	if !ok {
		return tools.ErrorResult("pid parameter is required and must be a number")
	}
	pid := int32(pidFloat)
	if pid <= 0 {
		return tools.ErrorResult(fmt.Sprintf("invalid pid %d: must be a positive integer", pid))
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("process %d not found: %v", pid, err))
	}

	result := SBOMResult{PID: pid}
	result.ProcessName, _ = p.NameWithContext(ctx)
	result.Executable, _ = p.ExeWithContext(ctx)

	switch runtime.GOOS {
	case "linux":
		result.Libraries = getLinuxLibraries(ctx, pid)
	case "darwin":
		result.Libraries = getDarwinLibraries(ctx, pid)
	case "windows":
		result.Libraries = getWindowsLibraries(ctx, pid)
	default:
		result.Error = fmt.Sprintf("SBOM inspection not supported on %s", runtime.GOOS)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return tools.NewToolResult(string(data))
}

func getLinuxLibraries(ctx context.Context, pid int32) []SBOMEntry {
	var libs []SBOMEntry

	// Read /proc/<pid>/maps to find loaded shared libraries
	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)
	data, err := os.ReadFile(mapsPath)
	if err != nil {
		slog.Warn("failed to read proc maps", "path", mapsPath, "error", err)
		return libs
	}

	seen := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		path := fields[len(fields)-1]
		if !strings.HasPrefix(path, "/") {
			continue
		}
		if seen[path] {
			continue
		}
		seen[path] = true

		if strings.HasSuffix(path, ".so") || strings.Contains(path, ".so.") {
			libs = append(libs, SBOMEntry{
				Library: filepath.Base(path),
				Path:    path,
				Type:    "shared_library",
			})
		}
	}

	// Also try ldd on the executable for static analysis
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	if cmd := exec.CommandContext(ctx, "ldd", exePath); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			for _, line := range strings.Split(string(output), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "linux-vdso") {
					continue
				}
				parts := strings.Split(line, "=>")
				if len(parts) == 2 {
					libPath := strings.TrimSpace(strings.Split(parts[1], "(")[0])
					libName := strings.TrimSpace(parts[0])
					if libPath != "" && !seen[libPath] {
						seen[libPath] = true
						libs = append(libs, SBOMEntry{
							Library: libName,
							Path:    libPath,
							Type:    "linked_library",
						})
					}
				}
			}
		}
	}

	return libs
}

func getDarwinLibraries(ctx context.Context, pid int32) []SBOMEntry {
	var libs []SBOMEntry
	cmd := exec.CommandContext(ctx, "vmmap", fmt.Sprintf("%d", pid))
	output, err := cmd.Output()
	if err != nil {
		slog.Warn("vmmap failed for SBOM inspection", "pid", pid, "error", err)
		return libs
	}

	seen := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, ".dylib") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.Contains(f, ".dylib") && !seen[f] {
					seen[f] = true
					libs = append(libs, SBOMEntry{
						Library: filepath.Base(f),
						Path:    f,
						Type:    "dylib",
					})
				}
			}
		}
	}
	return libs
}

func getWindowsLibraries(ctx context.Context, pid int32) []SBOMEntry {
	var libs []SBOMEntry
	// Use tasklist to get loaded DLLs
	cmd := exec.CommandContext(ctx, "tasklist", "/M", "/FI", fmt.Sprintf("PID eq %d", pid))
	output, err := cmd.Output()
	if err != nil {
		slog.Warn("tasklist failed for SBOM inspection", "pid", pid, "error", err)
		return libs
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(strings.ToLower(line), ".dll") {
			libs = append(libs, SBOMEntry{
				Library: line,
				Path:    line,
				Type:    "dll",
			})
		}
	}
	return libs
}
