package optimizer

import (
	"context"
	"fmt"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// OptimizeReport is the full optimization result metadata.
type OptimizeReport struct {
	InputTokens       int           `json:"input_tokens"`
	OutputTokens      int           `json:"output_tokens"`
	ReductionPercent  float64       `json:"reduction_percent"`
	TargetProfile     string        `json:"target_profile"`
	AppliedRules      []AppliedRule `json:"applied_rules"`
	TruncatedSections []string      `json:"truncated_sections"`
	StageResults      []StageResult `json:"stage_results"`
}

// Pipeline runs an ordered list of stages.
type Pipeline struct {
	Name    string
	Stages  []Stage
	Counter TokenCounter
}

// Run executes all stages and aggregates results.
func (pl *Pipeline) Run(ctx context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, OptimizeReport, error) {
	if p == nil {
		return nil, OptimizeReport{}, fmt.Errorf("pipeline %s: nil prompt", pl.Name)
	}

	inputTokens, err := pl.count(p)
	if err != nil {
		return nil, OptimizeReport{}, fmt.Errorf("count input tokens: %w", err)
	}

	report := OptimizeReport{
		InputTokens:   inputTokens,
		TargetProfile: string(cfg.TargetProfile),
	}

	current := p.Clone()
	var allRules []AppliedRule
	var allTruncated []string

	for _, stage := range pl.Stages {
		select {
		case <-ctx.Done():
			return nil, report, ctx.Err()
		default:
		}

		before, err := pl.count(current)
		if err != nil {
			return nil, report, err
		}

		next, result, err := stage.Apply(ctx, current, cfg)
		if err != nil {
			return nil, report, fmt.Errorf("stage %s: %w", stage.Name(), err)
		}
		current = next

		after, err := pl.count(current)
		if err != nil {
			return nil, report, err
		}
		result.TokensBefore = before
		result.TokensAfter = after
		if result.StageName == "" {
			result.StageName = stage.Name()
		}

		report.StageResults = append(report.StageResults, result)
		allRules = append(allRules, result.AppliedRules...)
		allTruncated = append(allTruncated, result.TruncatedSections...)
	}

	outputTokens, err := pl.count(current)
	if err != nil {
		return nil, report, err
	}

	report.OutputTokens = outputTokens
	report.AppliedRules = allRules
	report.TruncatedSections = uniqueStrings(allTruncated)
	if inputTokens > 0 {
		report.ReductionPercent = float64(inputTokens-outputTokens) / float64(inputTokens) * 100
	}

	return current, report, nil
}

func (pl *Pipeline) count(p *prompt.Prompt) (int, error) {
	if pl.Counter == nil {
		return len([]rune(p.Assemble())), nil
	}
	return pl.Counter.Count(p.Assemble())
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
