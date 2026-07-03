package bestpractice_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
)

func TestRulesForRefinement(t *testing.T) {
	t.Parallel()

	loader, err := infraBP.NewLoader()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := loader.Load(budget.ProfileCursor)
	if err != nil {
		t.Fatal(err)
	}

	rules := tp.RulesForRefinement()
	if len(rules) == 0 {
		t.Fatal("expected refinement rules")
	}
	found := false
	for _, r := range rules {
		if r.ID == "cursor-actionable-semantic" {
			found = true
			if r.Automatable {
				t.Fatal("refinement rule must not be automatable")
			}
		}
	}
	if !found {
		t.Fatal("cursor-actionable-semantic not found")
	}
}

func TestRule_LLMMaxOutputTokens(t *testing.T) {
	t.Parallel()

	r := bestpractice.Rule{LLM: bestpractice.RuleLLM{MaxOutputTokens: 512}}
	if r.LLMMaxOutputTokens() != 512 {
		t.Fatalf("got %d want 512", r.LLMMaxOutputTokens())
	}
}
