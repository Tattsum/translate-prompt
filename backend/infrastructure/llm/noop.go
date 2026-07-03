package llm

import (
	"context"
	"sync"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// NoopCompleter returns deterministic responses for tests and LLM-disabled mode.
type NoopCompleter struct {
	mu       sync.Mutex
	calls    int
	Response domainllm.CompletionOutcome
}

// NewNoopCompleter creates a noop completer with empty default response.
func NewNoopCompleter() *NoopCompleter {
	return &NoopCompleter{}
}

// Complete records the call and returns the configured response.
func (n *NoopCompleter) Complete(_ context.Context, intent domainllm.CompletionIntent, _ domainllm.CompletionBudget) (domainllm.CompletionOutcome, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++

	out := n.Response
	if out.Provider == "" {
		out.Provider = string(ProviderNoop)
	}
	if out.Content == "" && intent.Purpose == domainllm.PurposeSectionRefine {
		out.Content = intent.InputContent
	}
	return out, nil
}

// Calls returns how many completions were invoked.
func (n *NoopCompleter) Calls() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls
}
