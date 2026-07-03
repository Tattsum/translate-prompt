package llm

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

func TestParseIntakeFindings_JSON(t *testing.T) {
	t.Parallel()

	raw := `{"findings":[{"id":"x1","category":"goal_unclear","severity":4,"summary":"Clarify success criteria."}]}`
	got := ParseIntakeFindings(raw)
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got))
	}
	if got[0].Summary != "Clarify success criteria." {
		t.Fatalf("summary: got %q", got[0].Summary)
	}
	if got[0].Source != "llm" {
		t.Fatalf("source: got %q", got[0].Source)
	}
}

func TestEnrichIntakeOutcome_ParsesFindings(t *testing.T) {
	t.Parallel()

	out := enrichIntakeOutcome(domainllm.CompletionIntent{
		Purpose:       domainllm.PurposeIntakeAnalyze,
		TargetProfile: budget.ProfileCodex,
	}, domainllm.CompletionOutcome{
		Content: `{"findings":[{"id":"parsed","category":"goal_unclear","severity":5,"summary":"Need deployment target."}]}`,
	})
	if len(out.Findings) != 1 {
		t.Fatalf("expected parsed finding, got %d", len(out.Findings))
	}
	if out.Findings[0].Summary != "Need deployment target." {
		t.Fatalf("summary: got %q", out.Findings[0].Summary)
	}
}

func TestService_RecordUsageEnforcesOutputBudget(t *testing.T) {
	t.Parallel()

	svc := &Service{Config: Config{Enabled: true, DefaultMaxCalls: 3}}
	b := domainllm.CompletionBudget{MaxCalls: 3, MaxOutputTokens: 100}

	if err := svc.reserveBudget(b, 50); err != nil {
		t.Fatalf("first reserve: %v", err)
	}
	svc.RecordUsage(domainllm.CompletionUsage{OutputTokens: 50})

	if err := svc.reserveBudget(b, 50); err != nil {
		t.Fatalf("second reserve: %v", err)
	}
	svc.RecordUsage(domainllm.CompletionUsage{OutputTokens: 50})

	if err := svc.reserveBudget(b, 1); err == nil {
		t.Fatal("expected output token budget exceeded")
	}
}
