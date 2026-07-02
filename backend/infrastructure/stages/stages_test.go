package stages_test

import (
	"context"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/stages"
)

func TestNormalizeWhitespace(t *testing.T) {
	t.Parallel()
	stage := stages.NormalizeWhitespace{}
	p := prompt.New("hello    world\n\n\n\nfoo")
	out, result, err := stage.Apply(context.Background(), p, budget.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if result.StageName != "NormalizeWhitespace" {
		t.Fatalf("stage name: %s", result.StageName)
	}
	if out.Assemble() != "hello world\n\nfoo" {
		t.Fatalf("got %q", out.Assemble())
	}
}

func TestDeduplicateExact(t *testing.T) {
	t.Parallel()
	stage := stages.DeduplicateExact{}
	p := &prompt.Prompt{
		Sections: []prompt.Section{
			{ID: "a", Type: prompt.SectionTypeRules, Content: "same"},
			{ID: "b", Type: prompt.SectionTypeRules, Content: "same"},
			{ID: "c", Type: prompt.SectionTypeTask, Content: "task"},
		},
	}
	out, _, err := stage.Apply(context.Background(), p, budget.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(out.Sections))
	}
}

func TestParseSections_MarkdownHeaders(t *testing.T) {
	t.Parallel()
	stage := stages.ParseSections{}
	p := prompt.New("## Goal\nDo something\n\n## Context\nBackground info")
	out, _, err := stage.Apply(context.Background(), p, budget.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Sections) < 2 {
		t.Fatalf("expected multiple sections, got %d", len(out.Sections))
	}
}

func TestRemoveBoilerplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		pattern string
		want    string
	}{
		{"japanese filler", "以下の通り実装する", "以下の通り", "実装する"},
		{"english please", "Please fix the bug", "Please ", "fix the bug"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stage := stages.RemoveBoilerplate{Patterns: []string{tt.pattern}}
			p := prompt.New(tt.input)
			out, _, err := stage.Apply(context.Background(), p, budget.DefaultConfig())
			if err != nil {
				t.Fatal(err)
			}
			if out.Assemble() != tt.want {
				t.Fatalf("got %q want %q", out.Assemble(), tt.want)
			}
		})
	}
}
