package stages

import (
	"context"
	"regexp"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	"github.com/Tattsum/translate-prompt/backend/domain/refine"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
)

// LLMRefinerStage applies non-automatable rules via the LLM completer.
type LLMRefinerStage struct {
	Completer llm.Completer
	Loader    *infraBP.Loader
	Counter   optimizer.TokenCounter
}

func (LLMRefinerStage) Name() string { return "LLMRefiner" }

func (s LLMRefinerStage) Apply(ctx context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	if s.Completer == nil || !cfg.LLMEnabled {
		return cp, optimizer.StageResult{StageName: "LLMRefiner"}, nil
	}

	tp, err := s.Loader.Load(cfg.TargetProfile)
	if err != nil {
		return nil, optimizer.StageResult{}, err
	}

	b := llm.CompletionBudgetFrom(cfg)
	var applied []optimizer.AppliedRule

	for _, rule := range tp.RulesForRefinement() {
		for i := range cp.Sections {
			if !ruleMatchesSection(rule, cp.Sections[i], cp, cfg) {
				continue
			}
			intent := llm.CompletionIntent{
				Purpose:       llm.PurposeSectionRefine,
				TargetProfile: cfg.TargetProfile,
				RuleRef:       rule.ID,
				SectionRef:    sectionRef(cp.Sections[i], i),
				InputContent:  cp.Sections[i].Content,
				Constraints:   constraintsFromRule(rule),
			}
			outcome, err := s.Completer.Complete(ctx, intent, b)
			if err != nil {
				continue
			}
			refined := refine.Outcome{
				Content: outcome.Content,
				Applied: true,
			}
			if intent.Constraints.MustNotIncreaseTokens && s.Counter != nil {
				before, err := s.Counter.Count(cp.Sections[i].Content)
				if err != nil {
					return nil, optimizer.StageResult{}, err
				}
				after, err := s.Counter.Count(outcome.Content)
				if err != nil {
					return nil, optimizer.StageResult{}, err
				}
				if after > before {
					refined.Applied = false
					refined.RejectReason = refine.RejectTokenIncrease
				}
			}
			if !refined.Applied {
				continue
			}
			cp.Sections[i].Content = refined.Content
			applied = append(applied, optimizer.AppliedRule{
				ID:        rule.ID,
				SourceURL: rule.SourceURL,
				Method:    "llm",
				Model:     outcome.Usage.Model,
			})
		}
	}

	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "LLMRefiner", AppliedRules: applied}, nil
}

func sectionRef(s prompt.Section, index int) llm.SectionRef {
	return llm.SectionRef{Index: index, ID: s.ID, Type: s.Type}
}

func constraintsFromRule(rule bestpractice.Rule) llm.CompletionConstraints {
	c := llm.CompletionConstraints{MaxOutputTokens: rule.LLMMaxOutputTokens()}
	if rule.Constraints["must_not_increase_tokens"] {
		c.MustNotIncreaseTokens = true
	}
	if rule.Constraints["preserve_structure"] {
		c.PreserveStructure = true
	}
	if c.MaxOutputTokens <= 0 {
		c.MaxOutputTokens = 1024
	}
	return c
}

func ruleMatchesSection(rule bestpractice.Rule, section prompt.Section, p *prompt.Prompt, cfg budget.Config) bool {
	if tag := rule.Condition["section_tag"]; tag != "" {
		matched := false
		if section.Metadata != nil && section.Metadata["xml_tag"] == tag {
			matched = true
		}
		if strings.EqualFold(string(section.Type), tag) {
			matched = true
		}
		if !matched {
			return false
		}
	}
	if st := rule.Condition["section_type"]; st != "" {
		if !strings.EqualFold(string(section.Type), st) {
			return false
		}
	}
	if rule.Condition["remaining_tokens_over_budget"] == "true" {
		if !isOverBudget(p, cfg) {
			return false
		}
	}
	if rule.Action == "make_actionable_semantic" {
		return hasVaguePatterns(section.Content)
	}
	if rule.Action == "rewrite_imperative_semantic" {
		return section.Type == prompt.SectionTypeTask && hasResidualImperative(section.Content)
	}
	return true
}

func isOverBudget(p *prompt.Prompt, cfg budget.Config) bool {
	totalChars := p.TotalChars()
	// Heuristic: treat chars as proxy when token counter unavailable in condition check.
	return totalChars > cfg.MaxTokens*4
}

var (
	vaguePatterns      = regexp.MustCompile(`(?i)(きれいに|clean code|best practices)`)
	residualImperative = regexp.MustCompile(`(してほしい|改善して|なんとかして)`)
)

func hasVaguePatterns(content string) bool {
	return vaguePatterns.MatchString(content)
}

func hasResidualImperative(content string) bool {
	return residualImperative.MatchString(content)
}
