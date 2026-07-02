package intake_test

import (
	"context"
	"testing"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
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
}
