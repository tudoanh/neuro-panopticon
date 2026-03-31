package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"neuropanopticon/backend/agent"
	"neuropanopticon/backend/config"
	"neuropanopticon/backend/logging"
	"neuropanopticon/backend/models"
	"neuropanopticon/backend/scanner"
)

// App struct serves as the bridge between the Go backend and the React frontend.
// All exported methods are automatically bound to the Wails runtime.
type App struct {
	ctx     context.Context
	agent   *agent.Agent
	cfg     *config.AppConfig
	scanner *scanner.Scanner
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts. Initializes the agent.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	logging.Setup()
	slog.Info("NeuroPanopticon starting")

	cfg, err := config.Load()
	if err != nil {
		slog.Warn("failed to load config, using defaults", "error", err)
		cfg = config.DefaultConfig()
	}
	a.cfg = cfg
	slog.Info("config loaded", "backend", cfg.LLM.Backend, "scan_interval", cfg.Security.ScanIntervalSeconds)

	ag, err := agent.New(cfg)
	if err != nil {
		slog.Error("failed to initialize agent", "error", err)
		return
	}
	a.agent = ag
	// Wire Wails event emitter for real-time tool indicators
	a.agent.SetEmitter(func(eventName string, data any) {
		wailsRuntime.EventsEmit(a.ctx, eventName, data)
	})
	slog.Info("agent initialized", "backend", cfg.LLM.Backend)

	// Start background security scanner
	a.scanner = scanner.New(cfg.Security.ScanIntervalSeconds)
	a.scanner.Start(ctx)
}

// SendMessage sends a user message to the AI agent and returns the response.
// This is the primary IPC method called from the React chat interface.
func (a *App) SendMessage(message string) *agent.ChatResponse {
	if a.agent == nil {
		return &agent.ChatResponse{
			Content: "Agent is not initialized. Please check your LLM configuration in Settings.",
		}
	}

	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()

	resp, err := a.agent.Chat(ctx, message)
	if err != nil {
		slog.Error("agent chat failed", "error", err)
		return &agent.ChatResponse{
			Content: fmt.Sprintf("Error: %v", err),
		}
	}
	slog.Info("chat response", "tool_calls", len(resp.ToolCalls))
	return resp
}

// ResetChat clears the agent's conversation history.
func (a *App) ResetChat() {
	if a.agent != nil {
		a.agent.Reset()
	}
}

// GetSystemStatus returns current system metrics for the dashboard.
func (a *App) GetSystemStatus() *models.SystemStatus {
	status := &models.SystemStatus{
		AgentReady: a.agent != nil && a.agent.IsReady(),
		LLMBackend: "disconnected",
	}

	if a.cfg != nil {
		status.LLMBackend = a.cfg.LLM.Backend
	}

	// CPU usage
	cpuPercent, err := cpu.PercentWithContext(a.ctx, time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		status.CPUPercent = cpuPercent[0]
	}

	// Memory usage
	vmem, err := mem.VirtualMemoryWithContext(a.ctx)
	if err == nil {
		status.MemoryUsed = float64(vmem.Used) / (1024 * 1024 * 1024)
		status.MemoryTotal = float64(vmem.Total) / (1024 * 1024 * 1024)
	}

	// Real security score from background scanner
	if a.scanner != nil {
		status.Score = a.scanner.Score()
	} else {
		status.Score = models.SecurityScore{
			Score:       -1, // -1 means not yet scanned
			MaxScore:    100,
			LastUpdated: time.Now(),
		}
	}

	return status
}

// GetFindings returns the latest security findings from the background scanner.
func (a *App) GetFindings() []models.Finding {
	if a.scanner == nil {
		return nil
	}
	return a.scanner.Findings()
}

// shutdown is called when the Wails app is closing.
func (a *App) shutdown(ctx context.Context) {
	slog.Info("NeuroPanopticon shutting down")
	if a.scanner != nil {
		a.scanner.Stop()
	}
}

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.AppConfig {
	if a.cfg == nil {
		return config.DefaultConfig()
	}
	return a.cfg
}

// SaveConfig saves updated configuration and reinitializes the agent.
func (a *App) SaveConfig(cfg *config.AppConfig) error {
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	a.cfg = cfg
	slog.Info("config saved", "backend", cfg.LLM.Backend)

	// Reinitialize agent with new config
	ag, err := agent.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to reinitialize agent: %w", err)
	}
	a.agent = ag
	a.agent.SetEmitter(func(eventName string, data any) {
		wailsRuntime.EventsEmit(a.ctx, eventName, data)
	})
	slog.Info("agent reinitialized after config change")
	return nil
}
