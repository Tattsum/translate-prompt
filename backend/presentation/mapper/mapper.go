package mapper

import (
	"context"
	"strings"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	translatepromptv1 "github.com/Tattsum/translate-prompt/backend/gen/translate_prompt/v1"
	"github.com/Tattsum/translate-prompt/backend/graph/model"
)

// ProfileFromGraphQL converts gql enum to domain profile.
func ProfileFromGraphQL(p model.TargetProfile) budget.TargetProfile {
	switch p {
	case model.TargetProfileClaude:
		return budget.ProfileClaude
	case model.TargetProfileOpenai:
		return budget.ProfileOpenAI
	case model.TargetProfileDevin:
		return budget.ProfileDevin
	case model.TargetProfileCursor:
		return budget.ProfileCursor
	default:
		return budget.ProfileCodex
	}
}

// ProfileFromString converts wire string to domain profile.
func ProfileFromString(s string) budget.TargetProfile {
	p, ok := budget.ParseProfile(strings.ToLower(s))
	if !ok {
		return budget.ProfileCodex
	}
	return p
}

// ConfigFromGraphQL maps GraphQL input to domain config.
func ConfigFromGraphQL(in *model.OptimizeConfigInput) budget.Config {
	if in == nil {
		return budget.DefaultConfig()
	}
	return budget.Config{
		MaxTokens:     in.MaxTokens,
		TargetProfile: ProfileFromGraphQL(in.TargetProfile),
		Tokenizer:     in.Tokenizer,
		DeepDive:      derefBool(in.DeepDive),
		WorkspacePath: derefString(in.WorkspacePath),
	}
}

// ConfigFromProto maps protobuf config to domain config.
func ConfigFromProto(in *translatepromptv1.OptimizeConfig) budget.Config {
	if in == nil {
		return budget.DefaultConfig()
	}
	return budget.Config{
		MaxTokens:     int(in.MaxTokens),
		TargetProfile: ProfileFromString(in.TargetProfile),
		Tokenizer:     in.Tokenizer,
		DeepDive:      in.DeepDive,
		WorkspacePath: in.WorkspacePath,
	}
}

// AnalyzeToGraphQL maps domain analyze result to GraphQL model.
func AnalyzeToGraphQL(r domainintake.AnalyzeResult) *model.AnalyzeResult {
	out := &model.AnalyzeResult{}
	if r.Prompt != "" {
		out.Prompt = &r.Prompt
	}
	switch r.Status {
	case domainintake.StatusNeedsInput:
		out.Status = model.AnalyzeStatusNeedsInput
	default:
		out.Status = model.AnalyzeStatusReady
	}
	for _, q := range r.Questions {
		ruleID := q.RuleID
		out.Questions = append(out.Questions, &model.Question{
			ID:     q.ID,
			Text:   q.Text,
			RuleID: &ruleID,
		})
	}
	return out
}

// InvestigateToGraphQL maps investigation result.
func InvestigateToGraphQL(r domainintake.InvestigationResult) *model.InvestigationResult {
	out := &model.InvestigationResult{
		SuggestedCommands: r.SuggestedCommands,
	}
	for _, f := range r.Files {
		out.Files = append(out.Files, &model.InvestigationFile{
			Path:           f.Path,
			SectionType:    string(f.SectionType),
			ContentPreview: f.ContentPreview,
		})
	}
	return out
}

// OptimizeToGraphQL maps optimize result.
func OptimizeToGraphQL(r optimize.Result) *model.OptimizeResult {
	out := &model.OptimizeResult{
		OptimizedPrompt: r.OptimizedPrompt,
		Artifacts:       &model.OptimizeArtifacts{},
		Report: &model.OptimizeReport{
			InputTokens:       r.Report.InputTokens,
			OutputTokens:      r.Report.OutputTokens,
			ReductionPercent:  r.Report.ReductionPercent,
			TargetProfile:     r.Report.TargetProfile,
			TruncatedSections: r.Report.TruncatedSections,
		},
	}
	for _, rule := range r.Report.AppliedRules {
		delta := rule.TokensDelta
		out.Report.AppliedRules = append(out.Report.AppliedRules, &model.AppliedRule{
			ID:          rule.ID,
			SourceURL:   rule.SourceURL,
			TokensDelta: &delta,
		})
	}
	for _, m := range r.Artifacts.CursorMDCSuggestions {
		out.Artifacts.CursorMdcSuggestions = append(out.Artifacts.CursorMdcSuggestions, &model.MdcSuggestion{
			Filename: m["filename"],
			Content:  m["content"],
		})
	}
	return out
}

// AnalyzeToProto maps domain analyze result to protobuf.
func AnalyzeToProto(r domainintake.AnalyzeResult) *translatepromptv1.AnalyzeResponse {
	out := &translatepromptv1.AnalyzeResponse{
		Status: string(r.Status),
		Prompt: r.Prompt,
	}
	for _, q := range r.Questions {
		out.Questions = append(out.Questions, &translatepromptv1.Question{
			Id:     q.ID,
			Text:   q.Text,
			RuleId: q.RuleID,
		})
	}
	return out
}

// InvestigateToProto maps investigation result.
func InvestigateToProto(r domainintake.InvestigationResult) *translatepromptv1.InvestigateResponse {
	out := &translatepromptv1.InvestigateResponse{
		SuggestedCommands: r.SuggestedCommands,
	}
	for _, f := range r.Files {
		out.Files = append(out.Files, &translatepromptv1.InvestigationFile{
			Path:           f.Path,
			SectionType:    string(f.SectionType),
			ContentPreview: f.ContentPreview,
		})
	}
	return out
}

// OptimizeToProto maps optimize result.
func OptimizeToProto(r optimize.Result) *translatepromptv1.OptimizeResponse {
	out := &translatepromptv1.OptimizeResponse{
		OptimizedPrompt: r.OptimizedPrompt,
		Artifacts:       &translatepromptv1.OptimizeArtifacts{},
		Report: &translatepromptv1.OptimizeReport{
			InputTokens:       int32(r.Report.InputTokens),
			OutputTokens:      int32(r.Report.OutputTokens),
			ReductionPercent:  r.Report.ReductionPercent,
			TargetProfile:     r.Report.TargetProfile,
			TruncatedSections: r.Report.TruncatedSections,
		},
	}
	for _, rule := range r.Report.AppliedRules {
		out.Report.AppliedRules = append(out.Report.AppliedRules, &translatepromptv1.AppliedRule{
			Id:          rule.ID,
			SourceUrl:   rule.SourceURL,
			TokensDelta: int32(rule.TokensDelta),
		})
	}
	for _, m := range r.Artifacts.CursorMDCSuggestions {
		out.Artifacts.CursorMdcSuggestions = append(out.Artifacts.CursorMdcSuggestions, &translatepromptv1.MdcSuggestion{
			Filename: m["filename"],
			Content:  m["content"],
		})
	}
	return out
}

// PrepareOptimizePrompt merges answers and workspace context.
func PrepareOptimizePrompt(
	ctx context.Context,
	intakeUC *appintake.UseCase,
	prompt string,
	cfg budget.Config,
	answers map[string]string,
) (string, budget.Config, error) {
	promptText := intakeUC.MergeAnswers(prompt, answers)
	if cfg.WorkspacePath == "" {
		return promptText, cfg, nil
	}
	inv, err := intakeUC.Investigate(ctx, cfg.WorkspacePath, cfg.TargetProfile)
	if err != nil {
		return promptText, cfg, err
	}
	promptText = appintake.MergeContext(promptText, inv)
	cfg.VerificationCommands = inv.SuggestedCommands
	return promptText, cfg, nil
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
