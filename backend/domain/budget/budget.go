package budget

import "slices"

import "github.com/Tattsum/translate-prompt/backend/domain/prompt"

// TargetProfile identifies the output format provider.
type TargetProfile string

const (
	ProfileClaude TargetProfile = "claude"
	ProfileCodex  TargetProfile = "codex"
	ProfileOpenAI TargetProfile = "openai"
	ProfileDevin  TargetProfile = "devin"
	ProfileCursor TargetProfile = "cursor"
)

// AllProfiles returns supported target profiles.
func AllProfiles() []TargetProfile {
	return []TargetProfile{
		ProfileClaude,
		ProfileCodex,
		ProfileOpenAI,
		ProfileDevin,
		ProfileCursor,
	}
}

// ParseProfile validates and returns a TargetProfile.
func ParseProfile(s string) (TargetProfile, bool) {
	p := TargetProfile(s)
	if slices.Contains(AllProfiles(), p) {
		return p, true
	}
	return ProfileCodex, false
}

// Config holds optimization and intake settings.
type Config struct {
	MaxTokens            int
	TargetProfile        TargetProfile
	Tokenizer            string
	DeepDive             bool
	WorkspacePath        string
	VerificationCommands []string
	OutputMode           string // e.g. "session_brief" for Devin
	LLMEnabled           bool
	LLMMaxCalls          int
	LLMModelGemini       string
	LLMModelAnthropic    string
}

// EnrichLLMDefaults merges server-side LLM defaults into a request config.
func EnrichLLMDefaults(cfg, defaults Config) Config {
	if defaults.LLMEnabled {
		cfg.LLMEnabled = true
	}
	if cfg.LLMMaxCalls == 0 && defaults.LLMMaxCalls > 0 {
		cfg.LLMMaxCalls = defaults.LLMMaxCalls
	}
	if cfg.LLMModelGemini == "" && defaults.LLMModelGemini != "" {
		cfg.LLMModelGemini = defaults.LLMModelGemini
	}
	if cfg.LLMModelAnthropic == "" && defaults.LLMModelAnthropic != "" {
		cfg.LLMModelAnthropic = defaults.LLMModelAnthropic
	}
	return cfg
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxTokens:         8000,
		TargetProfile:     ProfileCodex,
		Tokenizer:         "cl100k_base",
		LLMMaxCalls:       3,
		LLMModelGemini:    "gemini-2.5-flash",
		LLMModelAnthropic: "claude-sonnet-4-20250514",
	}
}

// TokenBudget tracks token limits and per-section allocations.
type TokenBudget struct {
	Max         int
	Used        int
	Allocations map[prompt.SectionType]int
}

// NewTokenBudget creates a budget with default section allocations.
func NewTokenBudget(max int) *TokenBudget {
	if max <= 0 {
		max = 8000
	}
	return &TokenBudget{
		Max: max,
		Allocations: map[prompt.SectionType]int{
			prompt.SectionTypeTask:    max / 4,
			prompt.SectionTypeCode:    max / 4,
			prompt.SectionTypeRules:   max / 6,
			prompt.SectionTypeSkills:  max / 8,
			prompt.SectionTypeHistory: max / 10,
			prompt.SectionTypeOther:   max / 10,
		},
	}
}

// Remaining returns tokens left in the budget.
func (b *TokenBudget) Remaining() int {
	if b == nil {
		return 0
	}
	return b.Max - b.Used
}
