package llm

import (
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// CompletionPurpose identifies why the LLM is being called.
type CompletionPurpose string

const (
	PurposeIntakeAnalyze CompletionPurpose = "intake_analyze"
	PurposeSectionRefine CompletionPurpose = "section_refine"
)

// SectionRef identifies a prompt section for LLM targeting and finding linkage.
type SectionRef struct {
	Index int
	ID    string
	Type  prompt.SectionType
}

// Equal reports whether two section refs refer to the same section.
func (r SectionRef) Equal(other SectionRef) bool {
	return r.Index == other.Index && r.ID == other.ID && r.Type == other.Type
}

// CompletionConstraints carries domain constraints for a single completion.
type CompletionConstraints struct {
	MustNotIncreaseTokens bool
	PreserveStructure     bool
	MaxOutputTokens       int
}

// ContextFinding is a lightweight finding snapshot for LLM context.
// Application maps domain/intake.Finding values into this type.
type ContextFinding struct {
	ID         string
	Category   string
	Severity   int
	SectionRef SectionRef
	RuleID     string
	Summary    string
	Source     string
}

// PromptSectionSummary is a section excerpt passed as LLM context during intake.
type PromptSectionSummary struct {
	Ref     SectionRef
	Content string
}

// CompletionContext carries supplemental data for intake analysis.
type CompletionContext struct {
	HeuristicFindings []ContextFinding
	PromptSections    []PromptSectionSummary
}

// CompletionIntent describes what the LLM should do without embedding prompt templates.
type CompletionIntent struct {
	Purpose       CompletionPurpose
	TargetProfile budget.TargetProfile
	RuleRef       string
	SectionRef    SectionRef
	InputContent  string
	Constraints   CompletionConstraints
	Context       CompletionContext
}
