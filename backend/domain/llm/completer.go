package llm

import (
	"context"
	"time"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
)

// CompletionBudget limits LLM usage within a single Analyze or Optimize request.
type CompletionBudget struct {
	MaxCalls        int
	MaxOutputTokens int
	TimeoutPerCall  time.Duration
}

// DefaultCompletionBudget derives LLM budget from the final prompt token budget.
func DefaultCompletionBudget(maxTokens int) CompletionBudget {
	maxOutput := 0
	if maxTokens > 0 {
		maxOutput = min(maxTokens*3/10, 4000)
	}
	return CompletionBudget{
		MaxCalls:        3,
		MaxOutputTokens: maxOutput,
		TimeoutPerCall:  30 * time.Second,
	}
}

// Completer executes a completion from domain intent. Implementations live in infrastructure.
type Completer interface {
	Complete(ctx context.Context, intent CompletionIntent, budget CompletionBudget) (CompletionOutcome, error)
}

// BudgetResetter clears per-request LLM call counters. Implemented by infrastructure completers.
type BudgetResetter interface {
	ResetBudget()
}

// CompletionBudgetFrom derives an LLM budget from optimize/analyze configuration.
func CompletionBudgetFrom(cfg budget.Config) CompletionBudget {
	b := DefaultCompletionBudget(cfg.MaxTokens)
	if cfg.LLMMaxCalls > 0 {
		b.MaxCalls = cfg.LLMMaxCalls
	}
	return b
}

// ResetBudgetIfSupported clears per-request counters when the completer supports it.
func ResetBudgetIfSupported(c Completer) {
	if r, ok := c.(BudgetResetter); ok {
		r.ResetBudget()
	}
}
