package refine

import "github.com/Tattsum/translate-prompt/backend/domain/llm"

// Intent describes a section-level rewrite request for LLM refinement.
type Intent struct {
	SectionRef   llm.SectionRef
	RuleRef      string
	InputContent string
	Constraints  llm.CompletionConstraints
}

// Outcome records whether a refinement was applied.
type Outcome struct {
	Content      string
	Applied      bool
	RejectReason string
}

// Reject reasons for refinement outcomes.
const (
	RejectTokenIncrease  = "token_increase"
	RejectParseError     = "parse_error"
	RejectBudgetExceeded = "budget_exceeded"
	RejectRefusal        = "refusal"
	RejectProviderError  = "provider_error"
)
