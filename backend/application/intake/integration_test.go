package intake_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
)

func TestAnalyze_Integration_HeuristicFixture(t *testing.T) {
	t.Parallel()

	raw := readIntakeFixture(t, "heuristic_ambiguous.prompt.md")
	loader := mustLoader(t)
	uc := appintake.NewUseCase(loader)

	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCodex,
		DeepDive:      true,
		LLMEnabled:    false,
	}
	result, err := uc.Analyze(context.Background(), raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domainintake.StatusNeedsInput {
		t.Fatalf("status: got %q want needs_input", result.Status)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected heuristic findings")
	}
	for _, f := range result.Findings {
		if f.Source != domainintake.FindingSourceHeuristic {
			t.Fatalf("finding %q source: got %q want heuristic", f.ID, f.Source)
		}
	}
	assertQuestionIDs(t, result.Questions, "goal", "scope", "acceptance")
}

func TestAnalyze_Integration_CompleteFixture(t *testing.T) {
	t.Parallel()

	raw := readIntakeFixture(t, "heuristic_complete.prompt.md")
	loader := mustLoader(t)
	uc := appintake.NewUseCase(loader)

	cfg := budget.DefaultConfig()
	cfg.DeepDive = true
	result, err := uc.Analyze(context.Background(), raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domainintake.StatusReady {
		t.Fatalf("status: got %q want ready (%d questions)", result.Status, len(result.Questions))
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings for complete prompt, got %d", len(result.Findings))
	}
}

func TestAnalyze_Integration_LLMFixture(t *testing.T) {
	t.Parallel()

	raw := readIntakeFixture(t, "heuristic_ambiguous.prompt.md")
	loader := mustLoader(t)

	noop := infrallm.NewNoopCompleter()
	noop.Response = domainllm.CompletionOutcome{
		Findings: []domainllm.ContextFinding{{
			ID: "llm-deploy", Category: "scope_missing", Severity: 5,
			Summary: "Which deployment environment should be targeted?",
			Source:  "llm",
		}},
		Provider: "noop",
	}
	uc := appintake.NewUseCase(loader).WithCompleter(noop)

	cfg := budget.Config{
		MaxTokens:     8000,
		TargetProfile: budget.ProfileCodex,
		DeepDive:      true,
		LLMEnabled:    true,
	}
	result, err := uc.Analyze(context.Background(), raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domainintake.StatusNeedsInput {
		t.Fatalf("status: got %q want needs_input", result.Status)
	}

	var llmFinding bool
	for _, f := range result.Findings {
		if f.Source == domainintake.FindingSourceLLM && f.Summary == "Which deployment environment should be targeted?" {
			llmFinding = true
		}
	}
	if !llmFinding {
		t.Fatal("expected merged LLM finding in findings")
	}

	foundQuestion := false
	for _, q := range result.Questions {
		if q.Text == "Which deployment environment should be targeted?" {
			foundQuestion = true
		}
	}
	if !foundQuestion {
		t.Fatal("expected LLM finding to surface as question")
	}
	if noop.Calls() != 1 {
		t.Fatalf("noop calls: got %d want 1", noop.Calls())
	}
}

func readIntakeFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "intake", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustLoader(t *testing.T) *infraBP.Loader {
	t.Helper()
	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	return loader
}

func assertQuestionIDs(t *testing.T, questions []domainintake.Question, want ...string) {
	t.Helper()
	got := make(map[string]struct{}, len(questions))
	for _, q := range questions {
		got[q.ID] = struct{}{}
	}
	for _, id := range want {
		if _, ok := got[id]; !ok {
			t.Fatalf("missing question id %q in %+v", id, questions)
		}
	}
}
