package graph

import (
	"context"
	"fmt"

	"github.com/Tattsum/translate-prompt/backend/graph/model"
	"github.com/Tattsum/translate-prompt/backend/presentation/mapper"
)

// Analyze is the resolver for the analyze field.
func (r *mutationResolver) Analyze(ctx context.Context, input model.AnalyzeInput) (*model.AnalyzeResult, error) {
	cfg := mapper.ConfigFromGraphQL(input.Config)
	result, err := r.IntakeUC.Analyze(ctx, input.Prompt, cfg)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}
	return mapper.AnalyzeToGraphQL(result), nil
}

// Investigate is the resolver for the investigate field.
func (r *mutationResolver) Investigate(ctx context.Context, input model.InvestigateInput) (*model.InvestigationResult, error) {
	profile := mapper.ProfileFromGraphQL(input.TargetProfile)
	result, err := r.IntakeUC.Investigate(ctx, input.WorkspacePath, profile)
	if err != nil {
		return nil, fmt.Errorf("investigate: %w", err)
	}
	return mapper.InvestigateToGraphQL(result), nil
}

// Optimize is the resolver for the optimize field.
func (r *mutationResolver) Optimize(ctx context.Context, input model.OptimizeInput) (*model.OptimizeResult, error) {
	cfg := mapper.ConfigFromGraphQL(input.Config)
	answers := input.Answers
	if answers == nil {
		answers = map[string]string{}
	}
	promptText, cfg, err := mapper.PrepareOptimizePrompt(ctx, r.IntakeUC, input.Prompt, cfg, answers)
	if err != nil {
		return nil, fmt.Errorf("prepare prompt: %w", err)
	}
	result, err := r.OptimizeUC.Optimize(ctx, promptText, cfg)
	if err != nil {
		return nil, fmt.Errorf("optimize: %w", err)
	}
	return mapper.OptimizeToGraphQL(result), nil
}

// Health is the resolver for the health field.
func (r *queryResolver) Health(ctx context.Context) (*model.Health, error) {
	return &model.Health{Status: "ok"}, nil
}

// Estimate is the resolver for the estimate field.
func (r *queryResolver) Estimate(ctx context.Context, text string, tokenizer string) (*model.EstimateResult, error) {
	tokens, err := r.OptimizeUC.Estimate(text, tokenizer)
	if err != nil {
		return nil, fmt.Errorf("estimate: %w", err)
	}
	return &model.EstimateResult{Tokens: tokens}, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
