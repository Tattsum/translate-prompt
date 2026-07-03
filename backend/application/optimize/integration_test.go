package optimize_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
)

// TestOptimize_Integration_LLMRefiner_NoOp exercises the compress pipeline end-to-end
// with a noop Completer. Prompts are pre-shaped Section trees representing post-format
// state so automatable format rules do not consume LLM refiner triggers.
func TestOptimize_Integration_LLMRefiner_NoOp(t *testing.T) {
	t.Parallel()

	loader := mustOptimizeLoader(t)

	cases := []struct {
		name    string
		profile budget.TargetProfile
		sections []prompt.Section
		ruleID  string
		maxTok  int
		shorter string
	}{
		{
			name:    "cursor-actionable-semantic",
			profile: budget.ProfileCursor,
			sections: []prompt.Section{{
				ID: "rules-0", Type: prompt.SectionTypeRules,
				Content: "- Write clean code and follow best practices",
			}},
			ruleID:  "cursor-actionable-semantic",
			maxTok:  8000,
			shorter: "Use explicit constraints instead of vague guidance.",
		},
		{
			name:    "claude-explicit-semantic",
			profile: budget.ProfileClaude,
			sections: []prompt.Section{{
				ID: "task-0", Type: prompt.SectionTypeTask,
				Content: "この認証バグを改善してほしい。",
			}},
			ruleID:  "claude-explicit-semantic",
			maxTok:  8000,
			shorter: "Fix the auth bug in pkg/auth/session.go.",
		},
		{
			name:    "common-example-summarize",
			profile: budget.ProfileCodex,
			sections: []prompt.Section{{
				ID: "examples-0", Type: prompt.SectionTypeCode,
				Content: readOptimizeFixture(t, "llm_common_examples_body.md"),
				Metadata: map[string]string{"xml_tag": "examples"},
			}},
			ruleID:  "common-example-summarize",
			maxTok:  80,
			shorter: "Examples: charge, handle failure, emit event.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			noop := newRuleAwareNoop(tc.ruleID, tc.shorter)
			uc, err := optimize.NewUseCase(loader, "cl100k_base")
			if err != nil {
				t.Fatal(err)
			}
			uc.WithCompleter(noop)

			cfg := budget.Config{
				MaxTokens:     tc.maxTok,
				TargetProfile: tc.profile,
				Tokenizer:     "cl100k_base",
				LLMEnabled:    true,
				LLMMaxCalls:   5,
			}

			compressPL, err := uc.ExportBuildCompressPipeline(cfg)
			if err != nil {
				t.Fatal(err)
			}

			p := &prompt.Prompt{Sections: tc.sections}
			_, stageResult, err := compressPL.Run(context.Background(), p, cfg)
			if err != nil {
				t.Fatal(err)
			}

			rule := findAppliedRule(stageResult.AppliedRules, tc.ruleID)
			if rule == nil {
				t.Fatalf("expected applied rule %q, got %+v", tc.ruleID, stageResult.AppliedRules)
			}
			if rule.Method != "llm" {
				t.Fatalf("rule %q method: got %q want llm", tc.ruleID, rule.Method)
			}
			if rule.Model == "" {
				t.Fatalf("rule %q missing model metadata", tc.ruleID)
			}
			if rule.SourceURL == "" {
				t.Fatalf("rule %q missing source_url", tc.ruleID)
			}
		})
	}
}

func TestOptimize_Integration_FormatThenCompress_LLMRefiner(t *testing.T) {
	t.Parallel()

	loader := mustOptimizeLoader(t)
	noop := newRuleAwareNoop("cursor-actionable-semantic", "Use explicit lint rules.")
	uc, err := optimize.NewUseCase(loader, "cl100k_base")
	if err != nil {
		t.Fatal(err)
	}
	uc.WithCompleter(noop)

	raw := readOptimizeFixture(t, "llm_cursor_rules.prompt.md")
	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCursor,
		Tokenizer:     "cl100k_base",
		LLMEnabled:    true,
	}

	result, err := uc.Optimize(context.Background(), raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Format-stage automatable rules may consume vague phrases before LLMRefiner runs.
	if findAppliedRule(result.Report.AppliedRules, "cursor-actionable-semantic") != nil {
		t.Log("full optimize applied LLM refiner rule")
		return
	}
	if findAppliedRule(result.Report.AppliedRules, "cursor-actionable") == nil {
		t.Fatal("expected format-stage cursor-actionable to run")
	}
}

func TestOptimize_Integration_LLMDisabled_Phase1Regression(t *testing.T) {
	t.Parallel()

	loader := mustOptimizeLoader(t)
	uc, err := optimize.NewUseCase(loader, "cl100k_base")
	if err != nil {
		t.Fatal(err)
	}

	noop := infrallm.NewNoopCompleter()
	noop.Response = domainllm.CompletionOutcome{Content: "should not apply"}
	uc.WithCompleter(noop)

	raw := readOptimizeFixture(t, "llm_cursor_rules.prompt.md")
	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCursor,
		Tokenizer:     "cl100k_base",
		LLMEnabled:    false,
	}

	result, err := uc.Optimize(context.Background(), raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if findAppliedRule(result.Report.AppliedRules, "cursor-actionable-semantic") != nil {
		t.Fatal("LLM refiner rule must not apply when LLM_ENABLED=false")
	}
	if noop.Calls() > 0 {
		t.Fatalf("noop must not be called, got %d calls", noop.Calls())
	}
}

type ruleAwareNoop struct {
	ruleID  string
	content string
	calls   int
}

func newRuleAwareNoop(ruleID, content string) *ruleAwareNoop {
	return &ruleAwareNoop{ruleID: ruleID, content: content}
}

func (n *ruleAwareNoop) Complete(_ context.Context, intent domainllm.CompletionIntent, _ domainllm.CompletionBudget) (domainllm.CompletionOutcome, error) {
	n.calls++
	if intent.RuleRef != n.ruleID {
		return domainllm.CompletionOutcome{Content: intent.InputContent, Provider: "noop"}, nil
	}
	return domainllm.CompletionOutcome{
		Content:  n.content,
		Provider: "noop",
		Usage:    domainllm.CompletionUsage{Model: "noop-integration"},
	}, nil
}

func readOptimizeFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "optimize", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustOptimizeLoader(t *testing.T) *infraBP.Loader {
	t.Helper()
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	return loader
}

func findAppliedRule(rules []optimizer.AppliedRule, id string) *optimizer.AppliedRule {
	for i := range rules {
		if rules[i].ID == id {
			return &rules[i]
		}
	}
	return nil
}
