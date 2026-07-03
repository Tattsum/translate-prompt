package stages

import (
	"context"
	"regexp"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// ParseSections splits raw prompt into typed sections.
type ParseSections struct{}

func (ParseSections) Name() string { return "ParseSections" }

var (
	headerRe    = regexp.MustCompile(`(?m)^#{1,3}\s+(.+)$`)
	codeBlockRe = regexp.MustCompile("(?s)```[\\w]*\\n(.*?)```")
)

func (ParseSections) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	raw := strings.TrimSpace(cp.Raw)
	if raw == "" {
		return cp, optimizer.StageResult{StageName: "ParseSections"}, nil
	}

	// XML-tagged content (Claude style).
	if matches := findXMLTags(raw); len(matches) > 0 {
		cp.Sections = nil
		for i, m := range matches {
			tag := strings.ToLower(m.tag)
			content := strings.TrimSpace(m.content)
			cp.Sections = append(cp.Sections, prompt.Section{
				ID:      tag + "-" + itoa(i),
				Type:    mapTagToType(tag),
				Content: content,
				Metadata: map[string]string{
					"xml_tag": tag,
				},
			})
		}
		return cp, optimizer.StageResult{StageName: "ParseSections"}, nil
	}

	// Markdown headers.
	if locs := headerRe.FindAllStringSubmatchIndex(raw, -1); len(locs) > 0 {
		cp.Sections = nil
		for i, loc := range locs {
			title := raw[loc[2]:loc[3]]
			start := loc[1]
			end := len(raw)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			body := strings.TrimSpace(raw[start:end])
			cp.Sections = append(cp.Sections, prompt.Section{
				ID:      slugify(title) + "-" + itoa(i),
				Type:    mapHeaderToType(title),
				Content: body,
			})
		}
		// Preamble before first header as Task.
		preamble := strings.TrimSpace(raw[:locs[0][0]])
		if preamble != "" {
			cp.Sections = append([]prompt.Section{{
				ID: "preamble-0", Type: prompt.SectionTypeTask, Content: preamble,
			}}, cp.Sections...)
		}
		return cp, optimizer.StageResult{StageName: "ParseSections"}, nil
	}

	// Code blocks become Code sections; remainder is Task.
	if blocks := codeBlockRe.FindAllStringSubmatch(raw, -1); len(blocks) > 0 {
		cp.Sections = nil
		remaining := raw
		for i, b := range blocks {
			full := b[0]
			content := strings.TrimSpace(b[1])
			remaining = strings.Replace(remaining, full, "", 1)
			cp.Sections = append(cp.Sections, prompt.Section{
				ID: "code-" + itoa(i), Type: prompt.SectionTypeCode, Content: content,
			})
		}
		if task := strings.TrimSpace(remaining); task != "" {
			cp.Sections = append([]prompt.Section{{
				ID: "task-0", Type: prompt.SectionTypeTask, Content: task,
			}}, cp.Sections...)
		}
		return cp, optimizer.StageResult{StageName: "ParseSections"}, nil
	}

	// Default: single task section.
	cp.Sections = []prompt.Section{{ID: "task-0", Type: prompt.SectionTypeTask, Content: raw}}
	return cp, optimizer.StageResult{StageName: "ParseSections"}, nil
}

func mapTagToType(tag string) prompt.SectionType {
	switch tag {
	case "task":
		return prompt.SectionTypeTask
	case "rules":
		return prompt.SectionTypeRules
	case "skills":
		return prompt.SectionTypeSkills
	case "examples", "code":
		return prompt.SectionTypeCode
	case "context", "history":
		return prompt.SectionTypeHistory
	case "constraints":
		return prompt.SectionTypeRules
	default:
		return prompt.SectionTypeOther
	}
}

func mapHeaderToType(title string) prompt.SectionType {
	lower := strings.ToLower(strings.TrimSpace(title))
	switch {
	case strings.Contains(lower, "goal"), strings.Contains(lower, "task"), strings.Contains(lower, "what"):
		return prompt.SectionTypeTask
	case strings.Contains(lower, "context"), strings.Contains(lower, "background"), strings.Contains(lower, "history"):
		return prompt.SectionTypeHistory
	case strings.Contains(lower, "rule"), strings.Contains(lower, "constraint"), strings.Contains(lower, "done"):
		return prompt.SectionTypeRules
	case strings.Contains(lower, "skill"):
		return prompt.SectionTypeSkills
	case strings.Contains(lower, "code"), strings.Contains(lower, "example"):
		return prompt.SectionTypeCode
	default:
		return prompt.SectionTypeOther
	}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type xmlMatch struct {
	tag     string
	content string
}

func findXMLTags(raw string) []xmlMatch {
	tagRe := regexp.MustCompile(`<(\w+)>`)
	var matches []xmlMatch
	for _, loc := range tagRe.FindAllStringSubmatchIndex(raw, -1) {
		tag := raw[loc[2]:loc[3]]
		openEnd := loc[1]
		closeTag := "</" + tag + ">"
		closeIdx := strings.Index(raw[openEnd:], closeTag)
		if closeIdx < 0 {
			continue
		}
		content := raw[openEnd : openEnd+closeIdx]
		matches = append(matches, xmlMatch{tag: tag, content: content})
	}
	return matches
}
