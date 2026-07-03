package intake

import (
	"sort"

	"github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// FindingSource identifies how a finding was produced.
type FindingSource string

const (
	FindingSourceHeuristic FindingSource = "heuristic"
	FindingSourceLLM       FindingSource = "llm"
)

// Finding is a structured ambiguity detected during analysis.
type Finding struct {
	ID         string
	Category   string
	Severity   int
	SectionRef llm.SectionRef
	RuleID     string
	Summary    string
	Source     FindingSource
}

// AnalysisReport aggregates findings and flow status.
type AnalysisReport struct {
	Findings []Finding
	Status   Status
}

// MergeFindings deduplicates heuristic and LLM findings, keeping higher severity.
func MergeFindings(heuristic, llmFindings []Finding) []Finding {
	merged := make(map[string]Finding, len(heuristic)+len(llmFindings))
	for _, f := range heuristic {
		merged[findingKey(f)] = f
	}
	for _, f := range llmFindings {
		key := findingKey(f)
		if existing, ok := merged[key]; ok {
			if f.Severity > existing.Severity {
				merged[key] = f
			}
			continue
		}
		merged[key] = f
	}
	out := make([]Finding, 0, len(merged))
	for _, f := range merged {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return out[i].Severity > out[j].Severity
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func findingKey(f Finding) string {
	return f.Category + "|" + f.SectionRef.ID + "|" + string(f.SectionRef.Type)
}

// QuestionsFromFindings converts findings into user-facing questions deterministically.
func QuestionsFromFindings(findings []Finding) []Question {
	if len(findings) == 0 {
		return nil
	}
	sorted := append([]Finding(nil), findings...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return sorted[i].Severity > sorted[j].Severity
		}
		return sorted[i].ID < sorted[j].ID
	})
	qs := make([]Question, 0, len(sorted))
	seen := make(map[string]struct{}, len(sorted))
	for _, f := range sorted {
		if f.Severity <= 0 {
			continue
		}
		id := f.ID
		if id == "" {
			id = f.Category
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		text := f.Summary
		if text == "" {
			text = defaultQuestionText(f.Category)
		}
		qs = append(qs, Question{
			ID:     id,
			Text:   text,
			RuleID: f.RuleID,
		})
	}
	return qs
}

func defaultQuestionText(category string) string {
	switch category {
	case "goal_unclear":
		return "成功条件・完了の定義は？"
	case "scope_missing":
		return "対象ファイル・モジュールの範囲は？"
	case "contradiction":
		return "矛盾する要求があります。どちらを優先しますか？"
	case "acceptance_missing":
		return "完了条件・検証コマンドは？"
	default:
		return "追加の明確化が必要です。"
	}
}

// ToContextFinding maps a Finding to llm.ContextFinding for completion intents.
func ToContextFinding(f Finding) llm.ContextFinding {
	return llm.ContextFinding{
		ID:         f.ID,
		Category:   f.Category,
		Severity:   f.Severity,
		SectionRef: f.SectionRef,
		RuleID:     f.RuleID,
		Summary:    f.Summary,
		Source:     string(f.Source),
	}
}

// FindingFromContext maps llm.ContextFinding back to a domain Finding.
func FindingFromContext(f llm.ContextFinding) Finding {
	return Finding{
		ID:         f.ID,
		Category:   f.Category,
		Severity:   f.Severity,
		SectionRef: f.SectionRef,
		RuleID:     f.RuleID,
		Summary:    f.Summary,
		Source:     FindingSource(f.Source),
	}
}
