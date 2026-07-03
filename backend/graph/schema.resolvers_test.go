package graph_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/intake"
	"github.com/Tattsum/translate-prompt/backend/graph"
	"github.com/Tattsum/translate-prompt/backend/graph/model"
)

func TestInvestigate_Disabled(t *testing.T) {
	t.Parallel()
	resolver := graph.NewResolver(nil, nil, false, budget.Config{})
	_, err := resolver.Mutation().Investigate(context.Background(), model.InvestigateInput{
		WorkspacePath: "/tmp",
		TargetProfile: model.TargetProfileCodex,
	})
	if !errors.Is(err, intake.ErrInvestigateDisabled) {
		t.Fatalf("expected ErrInvestigateDisabled, got %v", err)
	}
}
