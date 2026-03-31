package agent

import (
	"fmt"

	"github.com/sipeed/picoclaw/pkg/providers"

	"neuropanopticon/backend/config"
)

// createProvider creates an LLM provider based on the app configuration.
// Supports Ollama (OpenAI-compatible API at localhost:11434) and cloud providers.
func createProvider(cfg *config.AppConfig) (providers.LLMProvider, string, error) {
	switch cfg.LLM.Backend {
	case "ollama":
		return createOllamaProvider(cfg)
	case "cloud":
		return createCloudProvider(cfg)
	default:
		return nil, "", fmt.Errorf("unknown LLM backend: %s", cfg.LLM.Backend)
	}
}

// createOllamaProvider creates an OpenAI-compatible provider pointing at a local Ollama instance.
// Ollama exposes an OpenAI-compatible API at /v1/chat/completions.
func createOllamaProvider(cfg *config.AppConfig) (providers.LLMProvider, string, error) {
	apiBase := cfg.LLM.OllamaURL + "/v1"
	model := cfg.LLM.OllamaModel
	if model == "" {
		model = "Qwen-3.5-9b"
	}

	provider := providers.NewHTTPProvider(
		"", // Ollama doesn't require an API key
		apiBase,
		"", // no proxy
	)

	return provider, model, nil
}

// createCloudProvider creates a provider for cloud-based LLM services.
func createCloudProvider(cfg *config.AppConfig) (providers.LLMProvider, string, error) {
	if cfg.LLM.CloudAPIKey == "" {
		return nil, "", fmt.Errorf("cloud API key is required for cloud backend")
	}

	apiBase := cfg.LLM.CloudURL
	if apiBase == "" {
		apiBase = "https://api.anthropic.com/v1"
	}

	model := cfg.LLM.CloudModel
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	provider := providers.NewHTTPProvider(
		cfg.LLM.CloudAPIKey,
		apiBase,
		"", // no proxy
	)

	return provider, model, nil
}
