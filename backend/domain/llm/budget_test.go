package llm_test

import (
	"testing"
	"time"

	"github.com/Tattsum/translate-prompt/backend/domain/llm"
)

func TestDefaultCompletionBudget(t *testing.T) {
	t.Parallel()

	b := llm.DefaultCompletionBudget(8000)
	if b.MaxCalls != 3 {
		t.Fatalf("MaxCalls: got %d want 3", b.MaxCalls)
	}
	if b.MaxOutputTokens != 2400 {
		t.Fatalf("MaxOutputTokens: got %d want 2400", b.MaxOutputTokens)
	}
	if b.TimeoutPerCall != 30*time.Second {
		t.Fatalf("TimeoutPerCall: got %v want 30s", b.TimeoutPerCall)
	}
}

func TestDefaultCompletionBudget_CapsAt4000(t *testing.T) {
	t.Parallel()

	b := llm.DefaultCompletionBudget(20000)
	if b.MaxOutputTokens != 4000 {
		t.Fatalf("MaxOutputTokens: got %d want 4000", b.MaxOutputTokens)
	}
}

func TestDefaultCompletionBudget_ZeroMaxTokens(t *testing.T) {
	t.Parallel()

	b := llm.DefaultCompletionBudget(0)
	if b.MaxOutputTokens != 0 {
		t.Fatalf("MaxOutputTokens: got %d want 0", b.MaxOutputTokens)
	}
}
