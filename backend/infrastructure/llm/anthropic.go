package llm

import (
	"context"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// AnthropicClient calls the Anthropic Messages API.
type AnthropicClient struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewAnthropicClient creates an Anthropic client.
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{
		client: client,
		model:  anthropic.Model(model),
	}
}

// Complete generates content from provider messages.
func (a *AnthropicClient) Complete(ctx context.Context, messages []ProviderMessage, maxOutput int) (domainllm.CompletionOutcome, error) {
	if a == nil || a.model == "" {
		return domainllm.CompletionOutcome{}, domainllm.ErrProviderUnavailable
	}

	var system []anthropic.TextBlockParam
	var anthropicMessages []anthropic.MessageParam
	for _, m := range messages {
		switch m.Role {
		case "system":
			system = append(system, anthropic.TextBlockParam{Text: m.Content})
		default:
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(
				anthropic.NewTextBlock(m.Content),
			))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: int64(maxOutputOrDefault(maxOutput)),
		Messages:  anthropicMessages,
	}
	if len(system) > 0 {
		params.System = system
	}

	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		if isRefusalFromAPIError(err) {
			return domainllm.CompletionOutcome{}, domainllm.ErrRefusal
		}
		return domainllm.CompletionOutcome{}, mapProviderError("anthropic", err)
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	content := strings.TrimSpace(text.String())
	if content == "" {
		return domainllm.CompletionOutcome{}, domainllm.ErrRefusal
	}

	return domainllm.CompletionOutcome{
		Content: content,
		Usage: domainllm.CompletionUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			Model:        string(a.model),
		},
		Provider: string(ProviderAnthropic),
	}, nil
}

func maxOutputOrDefault(maxOutput int) int64 {
	if maxOutput <= 0 {
		return 1024
	}
	return int64(maxOutput)
}
