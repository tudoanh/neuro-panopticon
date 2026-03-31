package config

import (
	"encoding/json"
	"fmt"
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
			OllamaURL:   "http://localhost:8999",
			OllamaModel: "Qwen-3.5-9b",
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

// Validate checks that all config values are within acceptable bounds.
// It returns an error describing the first invalid field found.
func (cfg *AppConfig) Validate() error {
	switch cfg.LLM.Backend {
	case "ollama", "cloud":
		// valid
	default:
		return fmt.Errorf("invalid llm.backend %q: must be \"ollama\" or \"cloud\"", cfg.LLM.Backend)
	}

	if cfg.LLM.Backend == "ollama" && cfg.LLM.OllamaURL == "" {
		return fmt.Errorf("llm.ollama_url must not be empty when backend is \"ollama\"")
	}

	if cfg.LLM.Backend == "cloud" && cfg.LLM.CloudAPIKey == "" {
		return fmt.Errorf("llm.cloud_api_key is required when backend is \"cloud\"")
	}

	if cfg.LLM.Temperature < 0 || cfg.LLM.Temperature > 2 {
		return fmt.Errorf("llm.temperature %.2f out of range [0, 2]", cfg.LLM.Temperature)
	}

	if cfg.LLM.MaxTokens < 1 || cfg.LLM.MaxTokens > 128000 {
		return fmt.Errorf("llm.max_tokens %d out of range [1, 128000]", cfg.LLM.MaxTokens)
	}

	if cfg.Security.ScanIntervalSeconds < 10 {
		return fmt.Errorf("security.scan_interval_seconds %d too low (minimum 10)", cfg.Security.ScanIntervalSeconds)
	}

	return nil
}

// ConfigDir returns the platform-appropriate config directory.
func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "NeuroPanopticon")
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
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk after validating it.
func Save(cfg *AppConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
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
