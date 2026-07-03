package llm

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
)

// Config holds LLM runtime settings loaded from environment variables.
type Config struct {
	Enabled         bool
	GeminiAPIKey    string
	AnthropicAPIKey string
	DefaultMaxCalls int
	GeminiModel     string
	AnthropicModel  string
	TimeoutPerCall  time.Duration
}

// LoadConfigFromEnv reads LLM configuration. API keys stay in infrastructure only.
func LoadConfigFromEnv() Config {
	enabled := strings.EqualFold(os.Getenv("LLM_ENABLED"), "true") || os.Getenv("LLM_ENABLED") == "1"

	maxCalls := 3
	if v := os.Getenv("LLM_DEFAULT_MAX_CALLS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxCalls = n
		}
	}

	geminiModel := os.Getenv("LLM_GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-2.5-flash"
	}

	anthropicModel := os.Getenv("LLM_ANTHROPIC_MODEL")
	if anthropicModel == "" {
		anthropicModel = "claude-sonnet-5"
	}

	return Config{
		Enabled:         enabled,
		GeminiAPIKey:    firstNonEmpty(os.Getenv("GOOGLE_API_KEY"), os.Getenv("GEMINI_API_KEY")),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DefaultMaxCalls: maxCalls,
		GeminiModel:     geminiModel,
		AnthropicModel:  anthropicModel,
		TimeoutPerCall:  30 * time.Second,
	}
}

// ApplyToBudgetConfig copies LLM settings into domain budget.Config.
func (c Config) ApplyToBudgetConfig(cfg *budget.Config) {
	cfg.LLMEnabled = c.Enabled
	if c.DefaultMaxCalls > 0 {
		cfg.LLMMaxCalls = c.DefaultMaxCalls
	}
	if c.GeminiModel != "" {
		cfg.LLMModelGemini = c.GeminiModel
	}
	if c.AnthropicModel != "" {
		cfg.LLMModelAnthropic = c.AnthropicModel
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
