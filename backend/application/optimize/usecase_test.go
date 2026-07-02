package optimize_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
)

func TestLoader_LoadAllProfiles(t *testing.T) {
	t.Parallel()
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range loader.ProfileNames() {
		tp, err := loader.Load(profile)
		if err != nil {
			t.Fatalf("load %s: %v", profile, err)
		}
		if len(tp.AllRules()) == 0 {
			t.Fatalf("profile %s has no rules", profile)
		}
	}
}

func TestOptimize_AllProfiles_Structure(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "verbose_prompt.md"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}

	profiles := []struct {
		name    budget.TargetProfile
		markers []string
	}{
		{budget.ProfileCodex, []string{"## Goal", "## Context", "## Constraints", "## Done when"}},
		{budget.ProfileClaude, []string{"<task>", "<context>", "<rules>", "<constraints>"}},
		{budget.ProfileOpenAI, []string{"## Goal", "## Done when"}},
		{budget.ProfileDevin, []string{"## What", "## How", "## Result"}},
		{budget.ProfileCursor, []string{}},
	}

	for _, tc := range profiles {
		t.Run(string(tc.name), func(t *testing.T) {
			t.Parallel()
			uc, err := optimize.NewUseCase(loader, "cl100k_base")
			if err != nil {
				t.Fatal(err)
			}
			cfg := budget.Config{
				MaxTokens:     8000,
				TargetProfile: tc.name,
				Tokenizer:     "cl100k_base",
			}
			result, err := uc.Optimize(context.Background(), string(raw), cfg)
			if err != nil {
				t.Fatal(err)
			}
			if result.Report.InputTokens <= 0 {
				t.Fatal("expected input tokens")
			}
			if len(result.Report.AppliedRules) == 0 {
				t.Fatal("expected applied rules with source_url")
			}
			for _, rule := range result.Report.AppliedRules {
				if rule.SourceURL == "" {
					t.Fatalf("rule %s missing source_url", rule.ID)
				}
			}
			for _, marker := range tc.markers {
				if !strings.Contains(result.OptimizedPrompt, marker) {
					t.Fatalf("profile %s missing marker %q in output:\n%s", tc.name, marker, result.OptimizedPrompt)
				}
			}
		})
	}
}

func TestOptimize_TokenReduction(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "verbose_prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	uc, err := optimize.NewUseCase(loader, "cl100k_base")
	if err != nil {
		t.Fatal(err)
	}
	cfg := budget.Config{
		MaxTokens:     200,
		TargetProfile: budget.ProfileCodex,
		Tokenizer:     "cl100k_base",
	}
	result, err := uc.Optimize(context.Background(), string(raw), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.ReductionPercent < 20 {
		t.Fatalf("expected >=20%% reduction, got %.1f%% (in=%d out=%d)",
			result.Report.ReductionPercent, result.Report.InputTokens, result.Report.OutputTokens)
	}
}
