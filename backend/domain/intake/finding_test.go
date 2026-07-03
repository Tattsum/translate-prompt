package intake_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

func TestMergeFindings_DedupeByCategoryAndSection(t *testing.T) {
	t.Parallel()

	ref := llm.SectionRef{ID: "task-0", Type: prompt.SectionTypeTask}
	heuristic := []intake.Finding{{
		ID: "h1", Category: "goal_unclear", Severity: 3, SectionRef: ref, Source: intake.FindingSourceHeuristic,
	}}
	llmFindings := []intake.Finding{{
		ID: "l1", Category: "goal_unclear", Severity: 5, SectionRef: ref,
		Summary: "What does success look like?", Source: intake.FindingSourceLLM,
	}}
	merged := intake.MergeFindings(heuristic, llmFindings)
	if len(merged) != 1 {
		t.Fatalf("got %d findings want 1", len(merged))
	}
	if merged[0].Severity != 5 {
		t.Fatalf("severity: got %d want 5", merged[0].Severity)
	}
	if merged[0].Source != intake.FindingSourceLLM {
		t.Fatalf("source: got %q want llm", merged[0].Source)
	}
}

func TestMergeFindings_KeepsDistinctSections(t *testing.T) {
	t.Parallel()

	heuristic := []intake.Finding{{
		ID: "h1", Category: "goal_unclear", Severity: 3,
		SectionRef: llm.SectionRef{ID: "task-0"},
	}}
	llmFindings := []intake.Finding{{
		ID: "l1", Category: "goal_unclear", Severity: 4,
		SectionRef: llm.SectionRef{ID: "rules-0"},
	}}
	merged := intake.MergeFindings(heuristic, llmFindings)
	if len(merged) != 2 {
		t.Fatalf("got %d findings want 2", len(merged))
	}
}

func TestQuestionsFromFindings_Deterministic(t *testing.T) {
	t.Parallel()

	findings := []intake.Finding{
		{ID: "b", Category: "scope_missing", Severity: 2, Summary: "Which files?"},
		{ID: "a", Category: "goal_unclear", Severity: 4, Summary: "Define success"},
	}
	qs := intake.QuestionsFromFindings(findings)
	if len(qs) != 2 {
		t.Fatalf("got %d questions want 2", len(qs))
	}
	if qs[0].ID != "a" || qs[0].Text != "Define success" {
		t.Fatalf("first question: %+v", qs[0])
	}
}

func TestQuestionsFromFindings_EmptySeveritySkipped(t *testing.T) {
	t.Parallel()

	qs := intake.QuestionsFromFindings([]intake.Finding{
		{ID: "x", Category: "goal_unclear", Severity: 0},
	})
	if len(qs) != 0 {
		t.Fatalf("got %d questions want 0", len(qs))
	}
}
