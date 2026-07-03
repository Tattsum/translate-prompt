package connectrpc_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/intake"
	translatepromptv1 "github.com/Tattsum/translate-prompt/backend/gen/translate_prompt/v1"
	"github.com/Tattsum/translate-prompt/backend/presentation/connectrpc"
)

func TestInvestigate_Disabled(t *testing.T) {
	t.Parallel()
	svc := connectrpc.NewService(nil, nil, false, budget.Config{})
	_, err := svc.Investigate(context.Background(), connect.NewRequest(&translatepromptv1.InvestigateRequest{
		WorkspacePath: "/tmp",
		TargetProfile: "codex",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodePermissionDenied {
		t.Fatalf("code = %v, want %v", connectErr.Code(), connect.CodePermissionDenied)
	}
	if !errors.Is(connectErr, intake.ErrInvestigateDisabled) {
		t.Fatalf("expected ErrInvestigateDisabled, got %v", err)
	}
}
