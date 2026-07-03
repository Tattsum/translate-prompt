package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// ProviderMessage is a provider-neutral chat message.
type ProviderMessage struct {
	Role    string
	Content string
}

// PromptBuilder assembles provider messages from domain intent without storing templates in domain.
type PromptBuilder struct{}

// Build returns dialect-specific messages for the routed provider.
func (PromptBuilder) Build(intent domainllm.CompletionIntent, provider Provider) []ProviderMessage {
	switch intent.Purpose {
	case domainllm.PurposeIntakeAnalyze:
		return buildIntakeMessages(intent, provider)
	case domainllm.PurposeSectionRefine:
		return buildRefineMessages(intent, provider)
	default:
		return []ProviderMessage{{Role: "user", Content: intent.InputContent}}
	}
}

func buildIntakeMessages(intent domainllm.CompletionIntent, provider Provider) []ProviderMessage {
	system := "You analyze prompts for ambiguities. Respond with JSON only: {\"findings\":[{\"id\":\"...\",\"category\":\"goal_unclear|scope_missing|contradiction|acceptance_missing|...\",\"severity\":1-5,\"summary\":\"...\"}]}. Do not include findings already listed under heuristic_findings unless you disagree or can add materially new detail."
	user := buildIntakeUserContent(intent)
	if provider == ProviderAnthropic {
		return []ProviderMessage{
			{Role: "user", Content: system + "\n\n" + user},
		}
	}
	return []ProviderMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

func buildIntakeUserContent(intent domainllm.CompletionIntent) string {
	var b strings.Builder
	b.WriteString("prompt:\n")
	b.WriteString(intent.InputContent)

	if len(intent.Context.HeuristicFindings) > 0 {
		b.WriteString("\n\nheuristic_findings:\n")
		if data, err := json.Marshal(intent.Context.HeuristicFindings); err == nil {
			b.Write(data)
		}
	}
	if len(intent.Context.PromptSections) > 0 {
		b.WriteString("\n\nprompt_sections:\n")
		for _, s := range intent.Context.PromptSections {
			fmt.Fprintf(&b, "- [%s/%s] %s\n", s.Ref.ID, s.Ref.Type, truncateForPrompt(s.Content, 500))
		}
	}
	return b.String()
}

func truncateForPrompt(content string, max int) string {
	content = strings.TrimSpace(content)
	if len(content) <= max {
		return content
	}
	return content[:max] + "..."
}

func buildRefineMessages(intent domainllm.CompletionIntent, provider Provider) []ProviderMessage {
	system := "Rewrite the section per the rule while preserving meaning."
	if intent.Constraints.MustNotIncreaseTokens {
		system += " Do not increase length."
	}
	if intent.Constraints.PreserveStructure {
		system += " Preserve headings and XML tags."
	}
	user := intent.InputContent
	if intent.RuleRef != "" {
		user = "rule_id: " + intent.RuleRef + "\n\n" + user
	}
	if provider == ProviderAnthropic {
		return []ProviderMessage{{Role: "user", Content: system + "\n\n" + user}}
	}
	return []ProviderMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}
