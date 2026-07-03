package llm_test

import (
	"context"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
)

func TestNoopCompleter_ReturnsConfiguredResponse(t *testing.T) {
	t.Parallel()

	noop := infrallm.NewNoopCompleter()
	noop.Response = domainllm.CompletionOutcome{
		Content:  "refined text",
		Provider: "noop",
	}
	out, err := noop.Complete(context.Background(), domainllm.CompletionIntent{
		Purpose:       domainllm.PurposeSectionRefine,
		TargetProfile: budget.ProfileCodex,
		InputContent:  "original",
	}, domainllm.DefaultCompletionBudget(8000))
	if err != nil {
		t.Fatal(err)
	}
	if out.Content != "refined text" {
		t.Fatalf("content: got %q", out.Content)
	}
	if noop.Calls() != 1 {
		t.Fatalf("calls: got %d want 1", noop.Calls())
	}
}

func TestRouteProvider(t *testing.T) {
	t.Parallel()

	if infrallm.RouteProvider(budget.ProfileClaude) != infrallm.ProviderAnthropic {
		t.Fatal("claude should route to anthropic")
	}
	if infrallm.RouteProvider(budget.ProfileCodex) != infrallm.ProviderGemini {
		t.Fatal("codex should route to gemini")
	}
}

func TestService_DisabledUsesNoop(t *testing.T) {
	t.Parallel()

	svc := infrallm.NewService(infrallm.Config{Enabled: false})
	out, err := svc.Complete(context.Background(), domainllm.CompletionIntent{
		Purpose:       domainllm.PurposeSectionRefine,
		TargetProfile: budget.ProfileCodex,
		InputContent:  "keep me",
	}, domainllm.DefaultCompletionBudget(8000))
	if err != nil {
		t.Fatal(err)
	}
	if out.Content != "keep me" {
		t.Fatalf("content: got %q", out.Content)
	}
}

func TestPromptBuilder_IntakeMessages(t *testing.T) {
	t.Parallel()

	builder := infrallm.PromptBuilder{}
	msgs := builder.Build(domainllm.CompletionIntent{
		Purpose:      domainllm.PurposeIntakeAnalyze,
		InputContent: "analyze this",
	}, infrallm.ProviderGemini)
	if len(msgs) < 2 {
		t.Fatalf("expected system+user, got %d messages", len(msgs))
	}
}
