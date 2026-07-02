package prompt

import (
	"maps"
	"strings"
	"unicode/utf8"
)

// SectionType classifies prompt fragments for truncation priority and formatting.
type SectionType string

const (
	SectionTypeTask    SectionType = "Task"
	SectionTypeRules   SectionType = "Rules"
	SectionTypeSkills  SectionType = "Skills"
	SectionTypeCode    SectionType = "Code"
	SectionTypeHistory SectionType = "History"
	SectionTypeOther   SectionType = "Other"
)

// DefaultPriority returns truncation priority (higher = keep longer).
func (t SectionType) DefaultPriority() int {
	switch t {
	case SectionTypeTask:
		return 100
	case SectionTypeCode:
		return 80
	case SectionTypeRules:
		return 60
	case SectionTypeSkills:
		return 50
	case SectionTypeHistory:
		return 20
	case SectionTypeOther:
		return 30
	default:
		return 30
	}
}

// Section is a classified fragment of a prompt.
type Section struct {
	ID       string
	Type     SectionType
	Content  string
	Priority int
	Metadata map[string]string
}

// PriorityValue returns effective priority for budget allocation.
func (s Section) PriorityValue() int {
	if s.Priority > 0 {
		return s.Priority
	}
	return s.Type.DefaultPriority()
}

// Prompt holds raw text and parsed sections plus optional artifacts.
type Prompt struct {
	Raw       string
	Sections  []Section
	Artifacts map[string]any
}

// New creates a Prompt from raw text with a single Task section.
func New(raw string) *Prompt {
	raw = strings.TrimSpace(raw)
	return &Prompt{
		Raw: raw,
		Sections: []Section{{
			ID:      "task-0",
			Type:    SectionTypeTask,
			Content: raw,
		}},
		Artifacts: make(map[string]any),
	}
}

// Clone returns a deep copy of the prompt.
func (p *Prompt) Clone() *Prompt {
	if p == nil {
		return nil
	}
	cp := &Prompt{
		Raw:       p.Raw,
		Sections:  make([]Section, len(p.Sections)),
		Artifacts: make(map[string]any, len(p.Artifacts)),
	}
	copy(cp.Sections, p.Sections)
	for i := range cp.Sections {
		if cp.Sections[i].Metadata != nil {
			md := make(map[string]string, len(cp.Sections[i].Metadata))
			maps.Copy(md, cp.Sections[i].Metadata)
			cp.Sections[i].Metadata = md
		}
	}
	maps.Copy(cp.Artifacts, p.Artifacts)
	return cp
}

// Assemble joins sections into a single string.
func (p *Prompt) Assemble() string {
	if len(p.Sections) == 0 {
		return strings.TrimSpace(p.Raw)
	}
	var b strings.Builder
	for i, s := range p.Sections {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(strings.TrimSpace(s.Content))
	}
	return strings.TrimSpace(b.String())
}

// TotalChars returns character count of assembled content.
func (p *Prompt) TotalChars() int {
	return utf8.RuneCountInString(p.Assemble())
}

// SectionsByType returns sections matching the given type.
func (p *Prompt) SectionsByType(t SectionType) []Section {
	var out []Section
	for _, s := range p.Sections {
		if s.Type == t {
			out = append(out, s)
		}
	}
	return out
}

// ContentByType concatenates section content for a type.
func (p *Prompt) ContentByType(t SectionType) string {
	var parts []string
	for _, s := range p.SectionsByType(t) {
		if c := strings.TrimSpace(s.Content); c != "" {
			parts = append(parts, c)
		}
	}
	return strings.Join(parts, "\n\n")
}
