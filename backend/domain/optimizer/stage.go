package optimizer

import (
	"context"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// AppliedRule records a rule application in the optimization report.
type AppliedRule struct {
	ID          string `json:"id"`
	SourceURL   string `json:"source_url"`
	TokensDelta int    `json:"tokens_delta"`
	Description string `json:"description,omitempty"`
}

// StageResult captures per-stage metrics.
type StageResult struct {
	StageName         string        `json:"stage_name"`
	TokensBefore      int           `json:"tokens_before"`
	TokensAfter       int           `json:"tokens_after"`
	AppliedRules      []AppliedRule `json:"applied_rules,omitempty"`
	TruncatedSections []string      `json:"truncated_sections,omitempty"`
}

// Stage transforms a prompt within the optimization pipeline.
type Stage interface {
	Name() string
	Apply(ctx context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, StageResult, error)
}

// TokenCounter counts tokens for reporting.
type TokenCounter interface {
	Count(text string) (int, error)
}
