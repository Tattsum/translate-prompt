package llm_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
)

func TestCompletionBudgetFrom_UsesLLMMaxCalls(t *testing.T) {
	t.Parallel()

	b := llm.CompletionBudgetFrom(budget.Config{MaxTokens: 1000, LLMMaxCalls: 7})
	if b.MaxCalls != 7 {
		t.Fatalf("max calls: got %d want 7", b.MaxCalls)
	}
}
