package llm

import "github.com/Tattsum/translate-prompt/backend/domain/budget"

// Provider identifies an LLM backend.
type Provider string

const (
	ProviderGemini    Provider = "gemini"
	ProviderAnthropic Provider = "anthropic"
	ProviderNoop      Provider = "noop"
)

// RouteProvider selects the LLM provider for a target profile.
func RouteProvider(profile budget.TargetProfile) Provider {
	if profile == budget.ProfileClaude {
		return ProviderAnthropic
	}
	return ProviderGemini
}
