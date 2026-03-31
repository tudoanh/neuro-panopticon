package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"neuropanopticon/backend/models"
	"neuropanopticon/backend/skills"
)

// Scanner runs periodic background security scans and maintains the current security posture.
type Scanner struct {
	mu sync.RWMutex

	interval time.Duration
	score    models.SecurityScore
	findings []models.Finding
	lastScan time.Time

	cancel context.CancelFunc
	done   chan struct{}
}

// New creates a new background Scanner with the given scan interval.
func New(intervalSeconds int) *Scanner {
	if intervalSeconds <= 0 {
		intervalSeconds = 300 // default 5 minutes
	}
	return &Scanner{
		interval: time.Duration(intervalSeconds) * time.Second,
		score: models.SecurityScore{
			Score:    -1, // not yet scanned
			MaxScore: 100,
		},
		done: make(chan struct{}),
	}
}

// Start begins the background scan loop. It runs an initial scan immediately,
// then repeats at the configured interval.
func (s *Scanner) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	slog.Info("scanner starting", "interval", s.interval)

	go func() {
		defer close(s.done)

		// Run initial scan
		s.runScan(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("scanner stopped")
				return
			case <-ticker.C:
				s.runScan(ctx)
			}
		}
	}()
}

// Stop halts the background scanner and waits for it to finish.
func (s *Scanner) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
}

// Score returns the most recent security score.
func (s *Scanner) Score() models.SecurityScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.score
}

// Findings returns the most recent findings list.
func (s *Scanner) Findings() []models.Finding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dst := make([]models.Finding, len(s.findings))
	copy(dst, s.findings)
	return dst
}

// LastScan returns the time of the last completed scan.
func (s *Scanner) LastScan() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastScan
}

// runScan performs one full scan cycle: network scan + lateral movement detection,
// then computes the security score from the aggregated findings.
func (s *Scanner) runScan(ctx context.Context) {
	start := time.Now()
	slog.Debug("scan cycle starting")

	var findings []models.Finding

	// 1. Network scan — check for risky listening ports
	netFindings := s.scanNetwork(ctx, start)
	findings = append(findings, netFindings...)

	// 2. Lateral movement detection — check process tree
	lmFindings := s.detectLateralMovement(ctx, start)
	findings = append(findings, lmFindings...)

	// 3. Blast radius context — exposed ports on all interfaces
	brFindings := s.checkBlastRadius(ctx, start)
	findings = append(findings, brFindings...)

	// Compute score
	score := ComputeScore(findings)

	s.mu.Lock()
	s.findings = findings
	s.score = score
	s.lastScan = start
	s.mu.Unlock()

	slog.Info("scan cycle complete",
		"duration", time.Since(start),
		"findings", len(findings),
		"score", score.Score,
		"critical", score.Critical,
		"high", score.High,
		"medium", score.Medium,
	)
}

// scanNetwork checks listening ports and flags risky ones as findings.
func (s *Scanner) scanNetwork(ctx context.Context, ts time.Time) []models.Finding {
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		slog.Error("network scan failed", "error", err)
		return nil
	}

	var findings []models.Finding
	seen := make(map[uint32]bool)

	for _, conn := range connections {
		if conn.Status != "LISTEN" {
			continue
		}
		port := conn.Laddr.Port
		if seen[port] {
			continue
		}
		seen[port] = true

		riskDesc, isRisky := skills.KnownRiskyPorts[port]
		if !isRisky {
			continue
		}

		procName := getProcessName(conn.Pid)
		severity := portSeverity(port)

		// Ports bound to 0.0.0.0 or :: are exposed to the network — higher risk
		exposed := conn.Laddr.IP == "0.0.0.0" || conn.Laddr.IP == "::" || conn.Laddr.IP == ""
		if exposed && severity == models.SeverityLow {
			severity = models.SeverityMedium
		}

		findings = append(findings, models.Finding{
			ID:          fmt.Sprintf("net-port-%d", port),
			Title:       fmt.Sprintf("Listening on port %d (%s)", port, riskDesc),
			Description: fmt.Sprintf("Process %q (PID %d) is listening on port %d. %s.", procName, conn.Pid, port, riskDesc),
			Severity:    severity,
			Category:    "network",
			Source:      "scan_network",
			Remediation: fmt.Sprintf("If this service is not needed, stop it or block port %d with a firewall rule.", port),
			Timestamp:   ts,
		})
	}
	return findings
}

// detectLateralMovement checks for suspicious process relationships and command patterns.
func (s *Scanner) detectLateralMovement(ctx context.Context, ts time.Time) []models.Finding {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		slog.Error("lateral movement scan failed", "error", err)
		return nil
	}

	var findings []models.Finding

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

		// Check suspicious parent-child
		if suspChildren, ok := skills.SuspiciousParentChild[toLower(parentName)]; ok {
			nameLower := toLower(name)
			for _, child := range suspChildren {
				if nameLower == child || contains(nameLower, child) {
					findings = append(findings, models.Finding{
						ID:          fmt.Sprintf("lm-pchild-%d", p.Pid),
						Title:       fmt.Sprintf("Suspicious child process: %s spawned by %s", name, parentName),
						Description: fmt.Sprintf("PID %d (%s) was spawned by %s (PID %d). This is a common lateral movement pattern.", p.Pid, name, parentName, ppid),
						Severity:    models.SeverityHigh,
						Category:    "lateral_movement",
						Source:      "detect_lateral_movement",
						Remediation: fmt.Sprintf("Investigate process %d and consider killing it if unauthorized.", p.Pid),
						Timestamp:   ts,
					})
					break
				}
			}
		}

		// Check suspicious command patterns
		cmdLower := toLower(cmdline)
		for _, pattern := range skills.SuspiciousCommandPatterns {
			if contains(cmdLower, toLower(pattern)) {
				findings = append(findings, models.Finding{
					ID:          fmt.Sprintf("lm-cmd-%d-%s", p.Pid, sanitizeID(pattern)),
					Title:       fmt.Sprintf("Suspicious command pattern: %s", pattern),
					Description: fmt.Sprintf("Process %s (PID %d) command line contains suspicious pattern %q.", name, p.Pid, pattern),
					Severity:    models.SeverityMedium,
					Category:    "lateral_movement",
					Source:      "detect_lateral_movement",
					Remediation: fmt.Sprintf("Investigate process %d (%s) and its parent %s.", p.Pid, name, parentName),
					Timestamp:   ts,
				})
				break
			}
		}
	}
	return findings
}

// checkBlastRadius checks for network-level exposure indicators.
func (s *Scanner) checkBlastRadius(ctx context.Context, ts time.Time) []models.Finding {
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		slog.Error("blast radius scan failed", "error", err)
		return nil
	}

	exposedCount := 0
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			if conn.Laddr.IP == "0.0.0.0" || conn.Laddr.IP == "::" || conn.Laddr.IP == "" {
				exposedCount++
			}
		}
	}

	var findings []models.Finding

	if exposedCount > 10 {
		findings = append(findings, models.Finding{
			ID:          "br-exposed-high",
			Title:       fmt.Sprintf("High number of exposed ports: %d", exposedCount),
			Description: fmt.Sprintf("%d ports are listening on all interfaces (0.0.0.0/::), significantly increasing attack surface.", exposedCount),
			Severity:    models.SeverityHigh,
			Category:    "blast_radius",
			Source:      "analyze_blast_radius",
			Remediation: "Review exposed services and bind them to localhost where possible.",
			Timestamp:   ts,
		})
	} else if exposedCount > 5 {
		findings = append(findings, models.Finding{
			ID:          "br-exposed-medium",
			Title:       fmt.Sprintf("Moderate number of exposed ports: %d", exposedCount),
			Description: fmt.Sprintf("%d ports are listening on all interfaces. Consider reducing exposure.", exposedCount),
			Severity:    models.SeverityMedium,
			Category:    "blast_radius",
			Source:      "analyze_blast_radius",
			Remediation: "Audit exposed services and restrict binding to specific interfaces.",
			Timestamp:   ts,
		})
	}

	return findings
}

// portSeverity returns the severity for a known risky port.
func portSeverity(port uint32) models.Severity {
	switch port {
	case 23: // telnet
		return models.SeverityCritical
	case 21, 135, 139, 445, 3389, 5900: // ftp, msrpc, netbios, smb, rdp, vnc
		return models.SeverityHigh
	case 6379, 27017, 9200: // redis, mongo, elasticsearch (often unauthenticated)
		return models.SeverityHigh
	case 22, 3306, 5432, 1433: // ssh, mysql, postgres, mssql
		return models.SeverityMedium
	case 80, 443, 8080, 8443, 53, 25: // web, dns, smtp
		return models.SeverityLow
	default:
		return models.SeverityInfo
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
	name, _ := p.Name()
	return name
}

func toLower(s string) string {
	// inline to avoid importing strings just for ToLower
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// sanitizeID makes a string safe for use in a finding ID.
func sanitizeID(s string) string {
	b := make([]byte, 0, len(s))
	for i := range s {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			b = append(b, c)
		} else if c >= 'A' && c <= 'Z' {
			b = append(b, c+'a'-'A')
		} else {
			b = append(b, '-')
		}
	}
	return string(b)
}

// MarshalFindings serializes findings to JSON for the frontend.
func MarshalFindings(findings []models.Finding) string {
	data, _ := json.MarshalIndent(findings, "", "  ")
	return string(data)
}
