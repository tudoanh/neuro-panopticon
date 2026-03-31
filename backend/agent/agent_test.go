package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/tools"

	"neuropanopticon/backend/config"
)

// slowProvider is a mock LLM provider that simulates a slow API call.
// It signals when the call starts, allowing tests to verify concurrent access.
type slowProvider struct {
	delay   time.Duration
	started chan struct{} // closed when Chat begins
}

func (p *slowProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	toolDefs []providers.ToolDefinition,
	model string,
	options map[string]any,
) (*providers.LLMResponse, error) {
	if p.started != nil {
		select {
		case <-p.started:
		default:
			close(p.started)
		}
	}
	select {
	case <-time.After(p.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return &providers.LLMResponse{
		Content:      "test response",
		FinishReason: "stop",
	}, nil
}

func (p *slowProvider) GetDefaultModel() string { return "test-model" }

// instantProvider returns immediately with a configurable response.
type instantProvider struct {
	response *providers.LLMResponse
}

func (p *instantProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	toolDefs []providers.ToolDefinition,
	model string,
	options map[string]any,
) (*providers.LLMResponse, error) {
	return p.response, nil
}

func (p *instantProvider) GetDefaultModel() string { return "test-model" }

func newTestAgent(provider providers.LLMProvider) *Agent {
	return &Agent{
		provider:      provider,
		tools:         tools.NewToolRegistry(),
		model:         "test-model",
		cfg:           config.DefaultConfig(),
		maxIterations: 10,
		messages: []providers.Message{
			{Role: "system", Content: "test system prompt"},
		},
	}
}

// TestResetNotBlockedDuringChat verifies that Reset() can execute while Chat()
// is waiting on a slow LLM API call. Before the fix, the mutex was held for
// the entire Chat() duration, causing Reset() to block.
func TestResetNotBlockedDuringChat(t *testing.T) {
	started := make(chan struct{})
	provider := &slowProvider{delay: 5 * time.Second, started: started}
	ag := newTestAgent(provider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Chat in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ag.Chat(ctx, "hello")
	}()

	// Wait for the LLM call to start
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for LLM call to start")
	}

	// Reset should complete quickly while Chat is blocked on LLM
	done := make(chan struct{})
	go func() {
		ag.Reset()
		close(done)
	}()

	select {
	case <-done:
		// Reset completed without blocking — test passes
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Reset() blocked for >500ms while Chat() was running — mutex held too long")
	}

	// Cancel the context to unblock the Chat goroutine
	cancel()
	wg.Wait()
}

// TestChatAddsMessages verifies Chat appends user and assistant messages correctly.
func TestChatAddsMessages(t *testing.T) {
	provider := &instantProvider{
		response: &providers.LLMResponse{
			Content:      "hello back",
			FinishReason: "stop",
		},
	}
	ag := newTestAgent(provider)

	resp, err := ag.Chat(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if resp.Content != "hello back" {
		t.Errorf("expected 'hello back', got %q", resp.Content)
	}

	// Should have: system + user + assistant = 3 messages
	ag.mu.Lock()
	n := len(ag.messages)
	ag.mu.Unlock()
	if n != 3 {
		t.Errorf("expected 3 messages (system+user+assistant), got %d", n)
	}
}

// TestResetClearsMessages verifies Reset restores only the system prompt.
func TestResetClearsMessages(t *testing.T) {
	provider := &instantProvider{
		response: &providers.LLMResponse{Content: "resp", FinishReason: "stop"},
	}
	ag := newTestAgent(provider)

	ag.Chat(context.Background(), "msg1")
	ag.Chat(context.Background(), "msg2")

	ag.Reset()

	ag.mu.Lock()
	n := len(ag.messages)
	role := ag.messages[0].Role
	ag.mu.Unlock()

	if n != 1 {
		t.Errorf("expected 1 message after reset, got %d", n)
	}
	if role != "system" {
		t.Errorf("expected system message after reset, got role %q", role)
	}
}
