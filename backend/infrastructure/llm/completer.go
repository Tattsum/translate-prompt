package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

// Service implements domainllm.Completer with routing, budgeting, and provider adapters.
type Service struct {
	Config    Config
	Builder   PromptBuilder
	Gemini    *GeminiClient
	Anthropic *AnthropicClient
	Noop      *NoopCompleter

	mu           sync.Mutex
	callsMade    int
	outputTokens int
}

// NewService wires provider clients from configuration.
func NewService(cfg Config) *Service {
	s := &Service{
		Config:  cfg,
		Builder: PromptBuilder{},
		Noop:    NewNoopCompleter(),
	}
	if cfg.GeminiAPIKey != "" {
		s.Gemini = NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
	}
	if cfg.AnthropicAPIKey != "" {
		s.Anthropic = NewAnthropicClient(cfg.AnthropicAPIKey, cfg.AnthropicModel)
	}
	return s
}

// Complete executes a completion respecting budget and provider routing.
func (s *Service) Complete(ctx context.Context, intent domainllm.CompletionIntent, b domainllm.CompletionBudget) (domainllm.CompletionOutcome, error) {
	if !s.Config.Enabled {
		return s.Noop.Complete(ctx, intent, b)
	}

	if err := s.reserveBudget(b, intent.Constraints.MaxOutputTokens); err != nil {
		return domainllm.CompletionOutcome{}, fmt.Errorf("llm budget purpose=%s: %w", intent.Purpose, err)
	}

	provider := RouteProvider(intent.TargetProfile)

	timeout := b.TimeoutPerCall
	if timeout <= 0 {
		timeout = s.Config.TimeoutPerCall
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	messages := s.Builder.Build(intent, provider)
	var (
		outcome domainllm.CompletionOutcome
		err     error
	)
	switch provider {
	case ProviderAnthropic:
		if s.Anthropic == nil {
			return domainllm.CompletionOutcome{}, fmt.Errorf("anthropic not configured: %w", domainllm.ErrProviderUnavailable)
		}
		outcome, err = s.Anthropic.Complete(callCtx, messages, intent.Constraints.MaxOutputTokens)
	case ProviderGemini:
		if s.Gemini == nil {
			return domainllm.CompletionOutcome{}, fmt.Errorf("gemini not configured: %w", domainllm.ErrProviderUnavailable)
		}
		outcome, err = s.Gemini.Complete(callCtx, messages, intent.Constraints.MaxOutputTokens)
	default:
		return domainllm.CompletionOutcome{}, fmt.Errorf("provider=%s purpose=%s: %w", provider, intent.Purpose, domainllm.ErrUnknownProvider)
	}
	if err != nil {
		return domainllm.CompletionOutcome{}, fmt.Errorf("llm complete provider=%s purpose=%s: %w", provider, intent.Purpose, err)
	}

	outcome = enrichIntakeOutcome(intent, outcome)
	s.RecordUsage(outcome.Usage)
	return outcome, nil
}

func enrichIntakeOutcome(intent domainllm.CompletionIntent, outcome domainllm.CompletionOutcome) domainllm.CompletionOutcome {
	if intent.Purpose == domainllm.PurposeIntakeAnalyze && len(outcome.Findings) == 0 {
		outcome.Findings = ParseIntakeFindings(outcome.Content)
	}
	return outcome
}

func (s *Service) reserveBudget(b domainllm.CompletionBudget, requestedOutput int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxCalls := b.MaxCalls
	if maxCalls <= 0 {
		maxCalls = s.Config.DefaultMaxCalls
	}
	if s.callsMade >= maxCalls {
		return domainllm.ErrBudgetExceeded
	}
	if requestedOutput > 0 && b.MaxOutputTokens > 0 && s.outputTokens+requestedOutput > b.MaxOutputTokens {
		return domainllm.ErrBudgetExceeded
	}
	s.callsMade++
	return nil
}

// BudgetFromConfig builds a domain completion budget from optimize/analyze config.
func BudgetFromConfig(cfg budget.Config) domainllm.CompletionBudget {
	return domainllm.CompletionBudgetFrom(cfg)
}

// ResetBudget clears per-request counters. Call at the start of Analyze/Optimize.
func (s *Service) ResetBudget() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callsMade = 0
	s.outputTokens = 0
}

// RecordUsage updates running output token totals after a successful call.
func (s *Service) RecordUsage(usage domainllm.CompletionUsage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outputTokens += usage.OutputTokens
}

// Ensure compile-time interface compliance.
var (
	_ domainllm.Completer      = (*Service)(nil)
	_ domainllm.BudgetResetter = (*Service)(nil)
	_ domainllm.Completer      = (*NoopCompleter)(nil)
)
