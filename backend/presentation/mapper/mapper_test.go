package mapper_test

import (
	"testing"

	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	"github.com/Tattsum/translate-prompt/backend/graph/model"
	"github.com/Tattsum/translate-prompt/backend/presentation/mapper"
)

func TestAnalyzeToGraphQL_IncludesFindings(t *testing.T) {
	t.Parallel()

	ruleID := "goal"
	sectionID := "task-0"
	sectionType := string(prompt.SectionTypeTask)
	out := mapper.AnalyzeToGraphQL(domainintake.AnalyzeResult{
		Status: domainintake.StatusNeedsInput,
		Findings: []domainintake.Finding{{
			ID: "goal", Category: "goal_unclear", Severity: 3,
			SectionRef: llm.SectionRef{ID: "task-0", Type: prompt.SectionTypeTask},
			RuleID: "goal", Summary: "Define success", Source: domainintake.FindingSourceHeuristic,
		}},
	})
	if len(out.Findings) != 1 {
		t.Fatalf("findings: got %d want 1", len(out.Findings))
	}
	f := out.Findings[0]
	if f.Source != model.FindingSourceHeuristic {
		t.Fatalf("source: got %v", f.Source)
	}
	if f.RuleID == nil || *f.RuleID != ruleID {
		t.Fatalf("ruleId: got %v want %q", f.RuleID, ruleID)
	}
	if f.SectionID == nil || *f.SectionID != sectionID {
		t.Fatalf("sectionId: got %v", f.SectionID)
	}
	if f.SectionType == nil || *f.SectionType != sectionType {
		t.Fatalf("sectionType: got %v", f.SectionType)
	}
}
