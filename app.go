package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"neuropanopticon/backend/agent"
	"neuropanopticon/backend/config"
	"neuropanopticon/backend/models"
)

// App struct serves as the bridge between the Go backend and the React frontend.
// All exported methods are automatically bound to the Wails runtime.
type App struct {
	ctx   context.Context
	agent *agent.Agent
	cfg   *config.AppConfig
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts. Initializes the agent.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: failed to load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}
	a.cfg = cfg

	ag, err := agent.New(cfg)
	if err != nil {
		fmt.Printf("Warning: failed to initialize agent: %v\n", err)
		return
	}
	a.agent = ag
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
		return &agent.ChatResponse{
			Content: fmt.Sprintf("Error: %v", err),
		}
	}
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

	// Placeholder security score
	status.Score = models.SecurityScore{
		Score:       -1, // -1 means not yet scanned
		MaxScore:    100,
		LastUpdated: time.Now(),
	}

	return status
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

	// Reinitialize agent with new config
	ag, err := agent.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to reinitialize agent: %w", err)
	}
	a.agent = ag
	return nil
}
