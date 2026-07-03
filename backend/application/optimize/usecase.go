package optimize

import (
	"context"
	"fmt"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/stages"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/tokenizer"
)

// Artifacts holds optional side outputs (e.g. cursor .mdc suggestions).
type Artifacts struct {
	CursorMDCSuggestions []map[string]string `json:"cursor_mdc_suggestions,omitempty"`
}

// Result is the full optimization output.
type Result struct {
	OptimizedPrompt string                   `json:"optimized_prompt"`
	Artifacts       Artifacts                `json:"artifacts"`
	Report          optimizer.OptimizeReport `json:"report"`
}

// UseCase orchestrates Format → Compress pipelines.
type UseCase struct {
	Loader    *infraBP.Loader
	Counter   optimizer.TokenCounter
	Completer llm.Completer
}

// NewUseCase creates an optimize use case with tokenizer.
func NewUseCase(loader *infraBP.Loader, encoding string) (*UseCase, error) {
	tok, err := tokenizer.New(encoding)
	if err != nil {
		return nil, err
	}
	return &UseCase{Loader: loader, Counter: tok}, nil
}

// WithCompleter attaches an LLM completer for refinement stages.
func (uc *UseCase) WithCompleter(c llm.Completer) *UseCase {
	uc.Completer = c
	return uc
}

// Optimize runs format then compress pipelines.
func (uc *UseCase) Optimize(ctx context.Context, raw string, cfg budget.Config) (Result, error) {
	if cfg.Tokenizer != "" {
		tok, err := tokenizer.New(cfg.Tokenizer)
		if err != nil {
			return Result{}, err
		}
		uc.Counter = tok
	}

	p := prompt.New(raw)
	llm.ResetBudgetIfSupported(uc.Completer)

	formatPL := &optimizer.Pipeline{
		Name:    "Format",
		Counter: uc.Counter,
		Stages: []optimizer.Stage{
			stages.ParseSections{},
			infraBP.ApplyBestPracticeProfile{Loader: uc.Loader},
			infraBP.ReorderOutcomeFirst{Loader: uc.Loader},
			infraBP.WrapStructure{Loader: uc.Loader, Counter: uc.Counter},
		},
	}

	formatted, formatReport, err := formatPL.Run(ctx, p, cfg)
	if err != nil {
		return Result{}, fmt.Errorf("format pipeline: %w", err)
	}

	compressPL, err := uc.buildCompressPipeline(cfg)
	if err != nil {
		return Result{}, err
	}

	compressed, compressReport, err := compressPL.Run(ctx, formatted, cfg)
	if err != nil {
		return Result{}, fmt.Errorf("compress pipeline: %w", err)
	}

	report := mergeReports(formatReport, compressReport)
	artifacts := extractArtifacts(compressed)

	return Result{
		OptimizedPrompt: compressed.Assemble(),
		Artifacts:       artifacts,
		Report:          report,
	}, nil
}

func (uc *UseCase) buildCompressPipeline(cfg budget.Config) (*optimizer.Pipeline, error) {
	tp, err := uc.Loader.Load(cfg.TargetProfile)
	if err != nil {
		return nil, err
	}

	var stageList []optimizer.Stage
	stageList = append(stageList,
		stages.NormalizeWhitespace{},
		stages.ASTParseStage{},
		stages.ASTCompressStage{},
		stages.DeduplicateExact{},
	)

	for _, rule := range tp.RulesForStage("RemoveBoilerplate") {
		switch rule.Action {
		case "remove_filler":
			stageList = append(stageList, stages.RemoveBoilerplate{
				Patterns: rule.Patterns,
				RuleID:   rule.ID,
				Source:   rule.SourceURL,
			})
		case "reduce_absolute_language":
			stageList = append(stageList, stages.ReduceAbsoluteLanguage{
				PreservePatterns: rule.PreservePatterns,
				RuleID:           rule.ID,
				Source:           rule.SourceURL,
			})
		case "remove_agent_known_facts":
			stageList = append(stageList, stages.RemoveAgentKnownFacts{
				RuleID: rule.ID,
				Source: rule.SourceURL,
			})
		}
	}

	stageList = append(stageList, stages.LLMRefinerStage{
		Completer: uc.Completer,
		Loader:    uc.Loader,
		Counter:   uc.Counter,
	})

	stageList = append(stageList,
		stages.BudgetAllocate{Counter: uc.Counter},
		stages.TruncateByPriority{Counter: uc.Counter},
	)

	for _, rule := range tp.RulesForStage("TruncateByPriority") {
		if rule.Action == "enforce_rules_budget" {
			maxWords := rule.Limits["max_words"]
			maxLines := rule.Limits["max_lines"]
			stageList = append(stageList, stages.EnforceRulesBudget{
				MaxWords: maxWords,
				MaxLines: maxLines,
				RuleID:   rule.ID,
				Source:   rule.SourceURL,
			})
		}
	}

	stageList = append(stageList, infraBP.AssembleWithProfile{})

	return &optimizer.Pipeline{
		Name:    "Compress",
		Counter: uc.Counter,
		Stages:  stageList,
	}, nil
}

func mergeReports(a, b optimizer.OptimizeReport) optimizer.OptimizeReport {
	out := a
	out.OutputTokens = b.OutputTokens
	out.ReductionPercent = 0
	if out.InputTokens > 0 {
		out.ReductionPercent = float64(out.InputTokens-out.OutputTokens) / float64(out.InputTokens) * 100
	}
	out.AppliedRules = append(out.AppliedRules, b.AppliedRules...)
	out.TruncatedSections = append(out.TruncatedSections, b.TruncatedSections...)
	out.StageResults = append(out.StageResults, b.StageResults...)
	return out
}

func extractArtifacts(p *prompt.Prompt) Artifacts {
	var art Artifacts
	if raw, ok := p.Artifacts["cursor_mdc_suggestions"]; ok {
		if items, ok := raw.([]map[string]string); ok {
			art.CursorMDCSuggestions = items
		}
	}
	return art
}

// Estimate counts tokens for text.
func (uc *UseCase) Estimate(text, encoding string) (int, error) {
	name := encoding
	if name == "" {
		name = "cl100k_base"
	}
	tok, err := tokenizer.New(name)
	if err != nil {
		return 0, err
	}
	return tok.Count(text)
}
