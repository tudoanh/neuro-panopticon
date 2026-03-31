package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// allowedActions defines the safe, predefined remediation actions.
var allowedActions = map[string]actionDef{
	"block_port": {
		Description: "Block inbound traffic on a specific port using the system firewall",
		RequiresTarget: true,
	},
	"kill_process": {
		Description: "Terminate a process by PID",
		RequiresTarget: true,
	},
	"disable_service": {
		Description: "Disable a systemd service (Linux only)",
		RequiresTarget: true,
	},
}

type actionDef struct {
	Description    string
	RequiresTarget bool
}

type RemediationResult struct {
	Action  string `json:"action"`
	Target  string `json:"target"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command_executed,omitempty"`
}

type RemediateTool struct {
	// Enabled controls whether remediation is actually allowed.
	Enabled            bool
	AllowedRemediations []string
}

func (t *RemediateTool) Name() string { return "remediate" }
func (t *RemediateTool) Description() string {
	return "Executes a safe, predefined OS command to remediate a security issue. " +
		"Supported actions: block_port (add firewall rule), kill_process (terminate by PID), " +
		"disable_service (stop a systemd service). " +
		"IMPORTANT: Only use this tool when the user explicitly approves the remediation action."
}

func (t *RemediateTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action_type": map[string]any{
				"type":        "string",
				"description": "The remediation action to perform: block_port, kill_process, or disable_service",
				"enum":        []string{"block_port", "kill_process", "disable_service"},
			},
			"target": map[string]any{
				"type":        "string",
				"description": "The target of the action (port number, PID, or service name)",
			},
		},
		"required": []string{"action_type", "target"},
	}
}

func (t *RemediateTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	if !t.Enabled {
		return tools.ErrorResult("Remediation is disabled in configuration. Enable it in Settings to allow the agent to take actions.")
	}

	actionType, _ := args["action_type"].(string)
	target, _ := args["target"].(string)

	if actionType == "" || target == "" {
		return tools.ErrorResult("Both action_type and target are required")
	}

	// Check if this action type is allowed
	if _, ok := allowedActions[actionType]; !ok {
		return tools.ErrorResult(fmt.Sprintf("Unknown action type: %s", actionType))
	}

	if len(t.AllowedRemediations) > 0 {
		allowed := false
		for _, a := range t.AllowedRemediations {
			if a == actionType {
				allowed = true
				break
			}
		}
		if !allowed {
			return tools.ErrorResult(fmt.Sprintf("Action %s is not in the allowed remediations list", actionType))
		}
	}

	// Sanitize target to prevent command injection
	target = sanitizeTarget(target)

	var result RemediationResult
	result.Action = actionType
	result.Target = target

	switch actionType {
	case "block_port":
		result = blockPort(ctx, target)
	case "kill_process":
		result = killProcess(ctx, target)
	case "disable_service":
		result = disableService(ctx, target)
	default:
		return tools.ErrorResult(fmt.Sprintf("Unhandled action type: %s", actionType))
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return tools.NewToolResult(string(data))
}

func blockPort(ctx context.Context, port string) RemediationResult {
	result := RemediationResult{Action: "block_port", Target: port}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "iptables", "-A", "INPUT", "-p", "tcp", "--dport", port, "-j", "DROP")
		result.Command = fmt.Sprintf("iptables -A INPUT -p tcp --dport %s -j DROP", port)
	case "darwin":
		// macOS uses pfctl
		result.Success = false
		result.Message = "macOS firewall management requires pfctl configuration (not yet implemented)"
		return result
	case "windows":
		cmd = exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "add", "rule",
			fmt.Sprintf("name=NeuroPanopticon_Block_%s", port),
			"dir=in", "action=block", fmt.Sprintf("localport=%s", port), "protocol=tcp")
		result.Command = fmt.Sprintf("netsh advfirewall firewall add rule name=NeuroPanopticon_Block_%s dir=in action=block localport=%s protocol=tcp", port, port)
	}

	if cmd != nil {
		if output, err := cmd.CombinedOutput(); err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed: %v - %s", err, string(output))
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("Successfully blocked port %s", port)
		}
	}

	return result
}

func killProcess(ctx context.Context, pid string) RemediationResult {
	result := RemediationResult{Action: "kill_process", Target: pid}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "taskkill", "/PID", pid, "/F")
		result.Command = fmt.Sprintf("taskkill /PID %s /F", pid)
	default:
		cmd = exec.CommandContext(ctx, "kill", "-9", pid)
		result.Command = fmt.Sprintf("kill -9 %s", pid)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed: %v - %s", err, string(output))
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("Successfully killed process %s", pid)
	}

	return result
}

func disableService(ctx context.Context, service string) RemediationResult {
	result := RemediationResult{Action: "disable_service", Target: service}

	if runtime.GOOS != "linux" {
		result.Success = false
		result.Message = "Service management is only supported on Linux (systemd)"
		return result
	}

	cmd := exec.CommandContext(ctx, "systemctl", "stop", service)
	result.Command = fmt.Sprintf("systemctl stop %s", service)

	if output, err := cmd.CombinedOutput(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed: %v - %s", err, string(output))
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("Successfully stopped service %s", service)
	}

	return result
}

// sanitizeTarget removes shell metacharacters to prevent injection.
func sanitizeTarget(s string) string {
	dangerous := []string{";", "|", "&", "`", "$", "(", ")", "{", "}", "<", ">", "\n", "\r"}
	for _, d := range dangerous {
		s = strings.ReplaceAll(s, d, "")
	}
	return strings.TrimSpace(s)
}
