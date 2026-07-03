package stages_test

import (
	"context"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/stages"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/tokenizer"
)

func TestLLMRefinerStage_RecordsAppliedRule(t *testing.T) {
	t.Parallel()

	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	tok, err := tokenizer.New("cl100k_base")
	if err != nil {
		t.Fatal(err)
	}

	noop := infrallm.NewNoopCompleter()
	noop.Response = domainllm.CompletionOutcome{
		Content:  "Use explicit constraints instead of best practices.",
		Provider: "noop",
		Usage:    domainllm.CompletionUsage{Model: "noop-model"},
	}

	stage := stages.LLMRefinerStage{
		Completer: noop,
		Loader:    loader,
		Counter:   tok,
	}

	p := &prompt.Prompt{
		Sections: []prompt.Section{{
			ID: "rules-0", Type: prompt.SectionTypeRules,
			Content: "Write clean code and follow best practices.",
		}},
	}
	cfg := budget.DefaultConfig()
	cfg.LLMEnabled = true
	cfg.TargetProfile = budget.ProfileCursor

	_, result, err := stage.Apply(context.Background(), p, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.AppliedRules) == 0 {
		t.Fatal("expected applied llm rule")
	}
	if result.AppliedRules[0].Method != "llm" {
		t.Fatalf("method: got %q want llm", result.AppliedRules[0].Method)
	}
	if result.AppliedRules[0].ID != "cursor-actionable-semantic" {
		t.Fatalf("rule id: got %q", result.AppliedRules[0].ID)
	}
	if noop.Calls() != 1 {
		t.Fatalf("expected one llm call, got %d", noop.Calls())
	}
}
