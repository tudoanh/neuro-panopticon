package config

import "testing"

func TestValidate_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig should be valid, got: %v", err)
	}
}

func TestValidate_InvalidBackend(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.Backend = "gpt-magic"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid backend")
	}
}

func TestValidate_EmptyOllamaURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.Backend = "ollama"
	cfg.LLM.OllamaURL = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty ollama URL")
	}
}

func TestValidate_CloudMissingKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.Backend = "cloud"
	cfg.LLM.CloudAPIKey = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing cloud API key")
	}
}

func TestValidate_CloudWithKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.Backend = "cloud"
	cfg.LLM.CloudAPIKey = "sk-test-key"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid cloud config should pass, got: %v", err)
	}
}

func TestValidate_TemperatureOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		temp float64
		ok   bool
	}{
		{"zero", 0, true},
		{"mid", 1.0, true},
		{"max", 2.0, true},
		{"negative", -0.1, false},
		{"over_max", 2.1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LLM.Temperature = tt.temp
			err := cfg.Validate()
			if tt.ok && err != nil {
				t.Fatalf("temperature %v should be valid, got: %v", tt.temp, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("temperature %v should be invalid", tt.temp)
			}
		})
	}
}

func TestValidate_MaxTokensBounds(t *testing.T) {
	tests := []struct {
		name   string
		tokens int
		ok     bool
	}{
		{"minimum", 1, true},
		{"typical", 4096, true},
		{"max", 128000, true},
		{"zero", 0, false},
		{"negative", -1, false},
		{"over_max", 128001, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LLM.MaxTokens = tt.tokens
			err := cfg.Validate()
			if tt.ok && err != nil {
				t.Fatalf("max_tokens %d should be valid, got: %v", tt.tokens, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("max_tokens %d should be invalid", tt.tokens)
			}
		})
	}
}

func TestValidate_ScanIntervalTooLow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Security.ScanIntervalSeconds = 5
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for scan interval below minimum")
	}

	cfg.Security.ScanIntervalSeconds = 10
	if err := cfg.Validate(); err != nil {
		t.Fatalf("scan interval 10 should be valid, got: %v", err)
	}
}
