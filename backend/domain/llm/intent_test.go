package llm_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

func TestSectionRef_Equal(t *testing.T) {
	t.Parallel()

	a := llm.SectionRef{Index: 1, ID: "task-0", Type: prompt.SectionTypeTask}
	b := llm.SectionRef{Index: 1, ID: "task-0", Type: prompt.SectionTypeTask}
	if !a.Equal(b) {
		t.Fatal("expected equal refs")
	}
	if a.Equal(llm.SectionRef{Index: 2, ID: "task-0", Type: prompt.SectionTypeTask}) {
		t.Fatal("expected unequal refs")
	}
}

func TestCompletionPurpose_Values(t *testing.T) {
	t.Parallel()

	if llm.PurposeIntakeAnalyze != "intake_analyze" {
		t.Fatalf("PurposeIntakeAnalyze: got %q", llm.PurposeIntakeAnalyze)
	}
	if llm.PurposeSectionRefine != "section_refine" {
		t.Fatalf("PurposeSectionRefine: got %q", llm.PurposeSectionRefine)
	}
}

func TestCompletionIntent_Fields(t *testing.T) {
	t.Parallel()

	intent := llm.CompletionIntent{
		Purpose:       llm.PurposeIntakeAnalyze,
		TargetProfile: budget.ProfileCodex,
		RuleRef:       "common-example-summarize",
		SectionRef:    llm.SectionRef{ID: "examples-0"},
		InputContent:  "example content",
		Constraints: llm.CompletionConstraints{
			MustNotIncreaseTokens: true,
			PreserveStructure:     true,
			MaxOutputTokens:       1024,
		},
	}
	if intent.Purpose != llm.PurposeIntakeAnalyze {
		t.Fatalf("unexpected purpose %q", intent.Purpose)
	}
	if intent.TargetProfile != budget.ProfileCodex {
		t.Fatalf("unexpected profile %q", intent.TargetProfile)
	}
}
