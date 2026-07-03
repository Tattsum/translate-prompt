package budget_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
)

func TestEnrichLLMDefaults(t *testing.T) {
	t.Parallel()

	defaults := budget.Config{
		LLMEnabled:        true,
		LLMMaxCalls:       5,
		LLMModelGemini:    "gemini-test",
		LLMModelAnthropic: "claude-test",
	}
	got := budget.EnrichLLMDefaults(budget.Config{MaxTokens: 4000}, defaults)
	if !got.LLMEnabled {
		t.Fatal("expected llm enabled")
	}
	if got.LLMMaxCalls != 5 {
		t.Fatalf("max calls: got %d", got.LLMMaxCalls)
	}
	if got.LLMModelGemini != "gemini-test" {
		t.Fatalf("gemini model: got %q", got.LLMModelGemini)
	}
}
