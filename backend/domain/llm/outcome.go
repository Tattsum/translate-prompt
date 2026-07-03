package llm

// CompletionUsage records token and model metadata for auditing.
type CompletionUsage struct {
	InputTokens  int
	OutputTokens int
	Model        string
}

// CompletionOutcome is the result of a single LLM completion.
type CompletionOutcome struct {
	Content  string
	Findings []ContextFinding
	Usage    CompletionUsage
	Provider string
}
