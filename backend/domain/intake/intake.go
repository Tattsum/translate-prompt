package intake

import "github.com/Tattsum/translate-prompt/backend/domain/prompt"

// Status represents intake flow state.
type Status string

const (
	StatusNeedsInput Status = "needs_input"
	StatusReady      Status = "ready"
)

// Question is shown when the prompt needs clarification.
type Question struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	RuleID string `json:"rule_id,omitempty"`
}

// Ambiguity describes a detected gap in the prompt.
type Ambiguity struct {
	Kind     string
	Question Question
	RuleID   string
}

// AnalyzeResult is returned by the intake analyzer.
type AnalyzeResult struct {
	Status    Status
	Questions []Question
	Prompt    string
	Findings  []Finding
}

// InvestigationFile is a workspace file discovered during investigation.
type InvestigationFile struct {
	Path           string             `json:"path"`
	SectionType    prompt.SectionType `json:"section_type"`
	ContentPreview string             `json:"content_preview"`
	Content        string             `json:"-"`
}

// InvestigationResult holds workspace scan output.
type InvestigationResult struct {
	Files             []InvestigationFile `json:"files"`
	SuggestedCommands []string            `json:"suggested_commands"`
}
