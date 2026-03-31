package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/sipeed/picoclaw/pkg/tools"
)

// SuspiciousParentChild defines known suspicious process parent-child relationships.
var SuspiciousParentChild = map[string][]string{
	// PDF readers should never spawn shells
	"evince":   {"bash", "sh", "zsh", "powershell", "cmd", "python", "perl", "curl", "wget", "nc", "ncat"},
	"okular":   {"bash", "sh", "zsh", "powershell", "cmd", "python", "perl", "curl", "wget", "nc", "ncat"},
	"acroread": {"bash", "sh", "zsh", "powershell", "cmd", "python", "perl", "curl", "wget", "nc", "ncat"},
	"mupdf":    {"bash", "sh", "zsh", "powershell", "cmd", "python", "perl", "curl", "wget", "nc", "ncat"},
	// Browsers spawning shells is suspicious
	"chrome":  {"bash", "sh", "powershell", "cmd", "python", "perl"},
	"firefox": {"bash", "sh", "powershell", "cmd", "python", "perl"},
	// Office apps spawning shells (Linux/macOS)
	"libreoffice": {"bash", "sh", "powershell", "cmd", "python", "perl", "curl", "wget"},
	"soffice":     {"bash", "sh", "powershell", "cmd", "python", "perl", "curl", "wget"},
	// Office apps spawning shells (Windows)
	"winword":  {"powershell", "cmd", "python", "perl", "curl", "certutil", "wscript", "cscript", "mshta"},
	"excel":    {"powershell", "cmd", "python", "perl", "curl", "certutil", "wscript", "cscript", "mshta"},
	"powerpnt": {"powershell", "cmd", "python", "perl", "curl", "certutil", "wscript", "cscript", "mshta"},
	"outlook":  {"powershell", "cmd", "python", "perl", "curl", "certutil", "wscript", "cscript", "mshta"},
	// macOS-specific apps
	"preview": {"bash", "sh", "zsh", "python", "perl", "curl", "wget", "nc"},
	"pages":   {"bash", "sh", "zsh", "python", "perl", "curl", "wget"},
}

// SuspiciousCommandPatterns are substrings that indicate lateral movement.
// Includes both Unix and Windows variants for cross-platform detection.
var SuspiciousCommandPatterns = []string{
	// Network reconnaissance
	"ping -c",   // Linux/macOS subnet scanning
	"ping -n",   // Windows subnet scanning
	"nmap",      // port scanning
	"nc -l",     // netcat listener
	"ncat -l",   // ncat listener
	"curl http",  // downloading payloads
	"wget http",  // downloading payloads
	"certutil -urlcache", // Windows payload download
	"bitsadmin /transfer", // Windows payload download

	// Credential access
	"/etc/passwd",          // Linux credential file
	"/etc/shadow",          // Linux credential file
	"sam /system",          // Windows SAM database
	"mimikatz",             // credential dumping tool
	"sekurlsa",             // mimikatz module

	// Persistence / execution
	"base64 -d",            // Unix encoded payloads
	"certutil -decode",     // Windows base64 decode
	"chmod +x",             // making downloaded files executable
	"ssh -o StrictHost",    // SSH with host key check disabled
	".ssh/authorized_keys", // SSH key injection

	// Firewall manipulation
	"iptables -F",          // Linux firewall flush
	"ufw disable",          // Ubuntu firewall disable
	"netsh advfirewall set", // Windows firewall manipulation
	"pfctl -d",             // macOS firewall disable
}

type LateralMovement struct {
	ProcessName string `json:"process_name"`
	PID         int32  `json:"pid"`
	ParentName  string `json:"parent_name"`
	ParentPID   int32  `json:"parent_pid"`
	Cmdline     string `json:"cmdline"`
	Reason      string `json:"reason"`
	Severity    string `json:"severity"`
}

type LateralMovementResult struct {
	Suspicious []LateralMovement `json:"suspicious_processes"`
	Scanned    int               `json:"processes_scanned"`
}

type LateralMovementDetectorTool struct{}

func (t *LateralMovementDetectorTool) Name() string { return "detect_lateral_movement" }
func (t *LateralMovementDetectorTool) Description() string {
	return "Analyzes running processes to detect lateral movement indicators. " +
		"Checks for suspicious parent-child process relationships (e.g., a PDF reader spawning a shell), " +
		"and suspicious command-line patterns that suggest reconnaissance or exploitation."
}

func (t *LateralMovementDetectorTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *LateralMovementDetectorTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to list processes: %v", err))
	}

	result := LateralMovementResult{Scanned: len(procs)}

	for _, p := range procs {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			continue
		}

		cmdline, _ := p.CmdlineWithContext(ctx)
		ppid, err := p.PpidWithContext(ctx)
		if err != nil {
			continue
		}

		parentName := ""
		if parent, err := process.NewProcess(ppid); err == nil {
			parentName, _ = parent.NameWithContext(ctx)
		}

		// Check suspicious parent-child relationships
		if suspiciousChildren, ok := SuspiciousParentChild[strings.ToLower(parentName)]; ok {
			nameLower := strings.ToLower(name)
			for _, child := range suspiciousChildren {
				if nameLower == child || strings.Contains(nameLower, child) {
					result.Suspicious = append(result.Suspicious, LateralMovement{
						ProcessName: name,
						PID:         p.Pid,
						ParentName:  parentName,
						ParentPID:   ppid,
						Cmdline:     truncate(cmdline, 200),
						Reason:      fmt.Sprintf("Suspicious child process: %s spawned by %s", name, parentName),
						Severity:    "high",
					})
				}
			}
		}

		// Check suspicious command patterns
		cmdLower := strings.ToLower(cmdline)
		for _, pattern := range SuspiciousCommandPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				result.Suspicious = append(result.Suspicious, LateralMovement{
					ProcessName: name,
					PID:         p.Pid,
					ParentName:  parentName,
					ParentPID:   ppid,
					Cmdline:     truncate(cmdline, 200),
					Reason:      fmt.Sprintf("Suspicious command pattern detected: %s", pattern),
					Severity:    "medium",
				})
				break
			}
		}
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return tools.NewToolResult(string(data))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
