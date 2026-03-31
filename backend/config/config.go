package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// AppConfig holds all application configuration.
type AppConfig struct {
	LLM      LLMConfig      `json:"llm"`
	Security SecurityConfig `json:"security"`
}

// LLMConfig configures the LLM backend.
type LLMConfig struct {
	// Backend is "ollama" (local) or "cloud" (Pro subscription).
	Backend string `json:"backend"`

	// Ollama settings (local LLM)
	OllamaURL   string `json:"ollama_url"`
	OllamaModel string `json:"ollama_model"`

	// Cloud settings (Panopti Pro)
	CloudAPIKey string `json:"cloud_api_key,omitempty"`
	CloudURL    string `json:"cloud_url,omitempty"`
	CloudModel  string `json:"cloud_model,omitempty"`

	// Shared settings
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	// AllowRemediation controls whether the agent can execute remediation actions.
	AllowRemediation bool `json:"allow_remediation"`

	// AllowedRemediations is the whitelist of allowed remediation action types.
	AllowedRemediations []string `json:"allowed_remediations"`

	// ScanIntervalSeconds is how often background scans run.
	ScanIntervalSeconds int `json:"scan_interval_seconds"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		LLM: LLMConfig{
			Backend:     "ollama",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "llama3.2",
			CloudURL:    "https://api.anthropic.com",
			CloudModel:  "claude-sonnet-4-6",
			MaxTokens:   4096,
			Temperature: 0.3,
		},
		Security: SecurityConfig{
			AllowRemediation:    false,
			AllowedRemediations: []string{"block_port", "kill_process"},
			ScanIntervalSeconds: 300,
		},
	}
}

// ConfigDir returns the platform-appropriate config directory.
func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "NeuroPanopticon")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "NeuroPanopticon")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "neuropanopticon")
	}
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load reads the config from disk, returning defaults if the file doesn't exist.
func Load() (*AppConfig, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg *AppConfig) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0o644)
}
