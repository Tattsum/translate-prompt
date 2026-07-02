package prompt_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

func TestPrompt_Assemble(t *testing.T) {
	t.Parallel()
	p := &prompt.Prompt{
		Sections: []prompt.Section{
			{Content: "first"},
			{Content: "second"},
		},
	}
	if got := p.Assemble(); got != "first\n\nsecond" {
		t.Fatalf("got %q", got)
	}
}

func TestSectionType_DefaultPriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		typ  prompt.SectionType
		want int
	}{
		{prompt.SectionTypeTask, 100},
		{prompt.SectionTypeHistory, 20},
	}
	for _, tt := range tests {
		if got := tt.typ.DefaultPriority(); got != tt.want {
			t.Fatalf("%s: got %d want %d", tt.typ, got, tt.want)
		}
	}
}
