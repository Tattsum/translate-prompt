package llm

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// GeminiClient calls the Gemini API via google.golang.org/genai.
type GeminiClient struct {
	client  *genai.Client
	model   string
	initErr error
}

// NewGeminiClient creates a Gemini client. apiKey may be empty for tests.
func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return &GeminiClient{model: model, initErr: err}
	}
	return &GeminiClient{client: client, model: model}
}

// Complete generates content from provider messages.
func (g *GeminiClient) Complete(ctx context.Context, messages []ProviderMessage, maxOutput int) (domainllm.CompletionOutcome, error) {
	if g.initErr != nil {
		return domainllm.CompletionOutcome{}, fmt.Errorf("gemini client init: %w", g.initErr)
	}
	if g.client == nil {
		return domainllm.CompletionOutcome{}, domainllm.ErrProviderUnavailable
	}

	var system string
	var parts []*genai.Part
	for _, m := range messages {
		switch m.Role {
		case "system":
			system = m.Content
		default:
			parts = append(parts, genai.NewPartFromText(m.Content))
		}
	}

	cfg := &genai.GenerateContentConfig{}
	if system != "" {
		cfg.SystemInstruction = genai.NewContentFromText(system, genai.RoleUser)
	}
	if maxOutput > 0 {
		cfg.MaxOutputTokens = int32(maxOutput)
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}, cfg)
	if err != nil {
		if isRefusalFromAPIError(err) {
			return domainllm.CompletionOutcome{}, domainllm.ErrRefusal
		}
		return domainllm.CompletionOutcome{}, mapProviderError("gemini", err)
	}
	text := strings.TrimSpace(resp.Text())
	if text == "" {
		return domainllm.CompletionOutcome{}, domainllm.ErrRefusal
	}

	usage := domainllm.CompletionUsage{Model: g.model}
	if resp.UsageMetadata != nil {
		usage.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return domainllm.CompletionOutcome{
		Content:  text,
		Usage:    usage,
		Provider: string(ProviderGemini),
	}, nil
}
