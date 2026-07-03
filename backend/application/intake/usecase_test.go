package intake_test

import (
	"context"
	"strings"
	"testing"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
)

func TestAnalyze_NeedsInput(t *testing.T) {
	t.Parallel()
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	uc := appintake.NewUseCase(loader)
	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCodex,
		DeepDive:      true,
	}
	result, err := uc.Analyze(context.Background(), "fix the bug", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "needs_input" {
		t.Fatalf("expected needs_input, got %s", result.Status)
	}
	if len(result.Questions) == 0 {
		t.Fatal("expected questions")
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings when deep_dive")
	}
}

func TestAnalyze_Ready(t *testing.T) {
	t.Parallel()
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	uc := appintake.NewUseCase(loader)
	cfg := budget.DefaultConfig()
	prompt := `## Goal
Fix auth bug in pkg/auth

## Scope
pkg/auth/*.go

## Acceptance
make test and make lint pass`
	result, err := uc.Analyze(context.Background(), prompt, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ready" {
		t.Fatalf("expected ready, got %s with %d questions", result.Status, len(result.Questions))
	}
}

func TestMergeAnswers(t *testing.T) {
	t.Parallel()
	out := appintake.NewUseCase(nil).MergeAnswers("base prompt", map[string]string{
		"goal": "tests pass",
	})
	if out == "base prompt" {
		t.Fatal("expected merged content")
	}
	if !strings.Contains(out, "## Clarifications") || !strings.Contains(out, "tests pass") {
		t.Fatalf("unexpected merge output: %q", out)
	}
}

func TestAnalyze_LLMComplement(t *testing.T) {
	t.Parallel()

	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	noop := infrallm.NewNoopCompleter()
	noop.Response = domainllm.CompletionOutcome{
		Findings: []domainllm.ContextFinding{{
			ID: "llm-extra", Category: "goal_unclear", Severity: 5,
			Summary: "Clarify the deployment target environment.",
			Source:  "llm",
		}},
	}
	uc := appintake.NewUseCase(loader).WithCompleter(noop)
	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCodex,
		DeepDive:      true,
		LLMEnabled:    true,
	}
	result, err := uc.Analyze(context.Background(), "fix the bug", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domainintake.StatusNeedsInput {
		t.Fatalf("expected needs_input, got %s", result.Status)
	}
	if noop.Calls() != 1 {
		t.Fatalf("expected one llm call, got %d", noop.Calls())
	}
	found := false
	for _, q := range result.Questions {
		if q.Text == "Clarify the deployment target environment." {
			found = true
		}
	}
	if !found {
		t.Fatal("expected LLM finding question")
	}
	for _, f := range result.Findings {
		if f.Source == domainintake.FindingSourceLLM && f.Summary == "Clarify the deployment target environment." {
			return
		}
	}
	t.Fatal("expected LLM finding in findings list")
}
