package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/tools"

	"neuropanopticon/backend/config"
	"neuropanopticon/backend/models"
	"neuropanopticon/backend/skills"
)

const systemPrompt = `You are NeuroPanopticon, an elite, locally-hosted cybersecurity auditor. Your mission is to protect this machine from lateral movement, supply chain attacks, and misconfigurations. You have access to specialized system tools (Skills). Do not guess system states; use your tools to fetch real-time data. If you detect an anomaly, explain the blast radius (how it could spread) and suggest a remediation tool. Prioritize user privacy: never upload sensitive file contents unless explicitly authorized via the Pro Cloud Fallback.

Available tools:
- scan_network: Scan listening ports and active connections
- detect_lateral_movement: Detect suspicious process activity
- analyze_blast_radius: Analyze network environment for threat scoring
- inspect_sbom: Analyze a process's loaded libraries (supply chain)
- remediate: Execute safe remediation actions (requires user approval)

When reporting findings, use severity levels: critical, high, medium, low, info.
Always explain WHY something is a risk and what the blast radius could be.`

// Agent orchestrates the AI agent loop for NeuroPanopticon.
type Agent struct {
	provider providers.LLMProvider
	tools    *tools.ToolRegistry
	model    string
	cfg      *config.AppConfig

	// Chat history per session
	mu       sync.Mutex
	messages []providers.Message

	maxIterations int
}

// New creates a new Agent with the given config.
func New(cfg *config.AppConfig) (*Agent, error) {
	provider, model, err := createProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating LLM provider: %w", err)
	}

	registry := tools.NewToolRegistry()
	skills.RegisterAll(registry, cfg)

	a := &Agent{
		provider:      provider,
		tools:         registry,
		model:         model,
		cfg:           cfg,
		maxIterations: 10,
	}

	// Initialize with system prompt
	a.messages = []providers.Message{
		{Role: "system", Content: systemPrompt},
	}

	return a, nil
}

// Chat sends a user message and runs the agent loop until a final response is produced.
// It returns the assistant's response and any tool calls that were made.
func (a *Agent) Chat(ctx context.Context, userMessage string) (*ChatResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add user message
	a.messages = append(a.messages, providers.Message{
		Role:    "user",
		Content: userMessage,
	})

	var allToolCalls []models.ToolCall

	// Agent loop: call LLM, handle tool calls, repeat until done
	for i := 0; i < a.maxIterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		toolDefs := a.tools.ToProviderDefs()
		resp, err := a.provider.Chat(ctx, a.messages, toolDefs, a.model, map[string]any{
			"temperature": a.cfg.LLM.Temperature,
			"max_tokens":  a.cfg.LLM.MaxTokens,
		})
		if err != nil {
			return nil, fmt.Errorf("LLM chat error: %w", err)
		}

		// If no tool calls, we have a final response
		if len(resp.ToolCalls) == 0 {
			a.messages = append(a.messages, providers.Message{
				Role:    "assistant",
				Content: resp.Content,
			})
			return &ChatResponse{
				Content:   resp.Content,
				ToolCalls: allToolCalls,
			}, nil
		}

		// Add assistant message with tool calls
		a.messages = append(a.messages, providers.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			toolName := tc.Name
			if tc.Function != nil {
				toolName = tc.Function.Name
			}

			args := tc.Arguments
			if tc.Function != nil && args == nil {
				// Parse function arguments from string
				args = make(map[string]any)
				json.Unmarshal([]byte(tc.Function.Arguments), &args)
			}

			start := time.Now()
			result := a.tools.Execute(ctx, toolName, args)
			duration := time.Since(start)

			status := "success"
			if result.IsError {
				status = "error"
			}

			tc := models.ToolCall{
				ID:       tc.ID,
				Name:     toolName,
				Result:   result.ContentForLLM(),
				Status:   status,
				Duration: duration.Milliseconds(),
			}
			argsJSON, _ := json.Marshal(args)
			tc.Args = string(argsJSON)
			allToolCalls = append(allToolCalls, tc)

			// Add tool result to conversation
			a.messages = append(a.messages, providers.Message{
				Role:       "tool",
				Content:    result.ContentForLLM(),
				ToolCallID: tc.ID,
			})
		}
	}

	return &ChatResponse{
		Content:   "I reached the maximum number of tool call iterations. Please try a more specific request.",
		ToolCalls: allToolCalls,
	}, nil
}

// Reset clears the conversation history.
func (a *Agent) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = []providers.Message{
		{Role: "system", Content: systemPrompt},
	}
}

// IsReady returns whether the agent's LLM provider is configured and accessible.
func (a *Agent) IsReady() bool {
	return a.provider != nil
}

// ChatResponse is the structured response from a Chat call.
type ChatResponse struct {
	Content   string            `json:"content"`
	ToolCalls []models.ToolCall `json:"tool_calls,omitempty"`
}
