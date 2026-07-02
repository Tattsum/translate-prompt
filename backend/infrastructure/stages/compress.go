package stages

import (
	"context"
	"regexp"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// NormalizeWhitespace collapses excessive whitespace.
type NormalizeWhitespace struct{}

func (NormalizeWhitespace) Name() string { return "NormalizeWhitespace" }

func (NormalizeWhitespace) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	spaceRe := regexp.MustCompile(`[ \t]+`)
	blankRe := regexp.MustCompile(`\n{3,}`)
	for i := range cp.Sections {
		content := cp.Sections[i].Content
		content = spaceRe.ReplaceAllString(content, " ")
		content = blankRe.ReplaceAllString(content, "\n\n")
		cp.Sections[i].Content = strings.TrimSpace(content)
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "NormalizeWhitespace"}, nil
}

// DeduplicateExact removes identical section content.
type DeduplicateExact struct{}

func (DeduplicateExact) Name() string { return "DeduplicateExact" }

func (DeduplicateExact) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	seen := make(map[string]struct{})
	var kept []prompt.Section
	for _, s := range cp.Sections {
		key := string(s.Type) + "\x00" + strings.TrimSpace(s.Content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		kept = append(kept, s)
	}
	cp.Sections = kept
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "DeduplicateExact"}, nil
}

// RemoveBoilerplate strips filler phrases from YAML patterns.
type RemoveBoilerplate struct {
	Patterns []string
	RuleID   string
	Source   string
}

func (r RemoveBoilerplate) Name() string { return "RemoveBoilerplate" }

func (r RemoveBoilerplate) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	var applied []optimizer.AppliedRule
	for i := range cp.Sections {
		before := cp.Sections[i].Content
		content := before
		for _, pat := range r.Patterns {
			if pat == "" {
				continue
			}
			re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pat))
			content = re.ReplaceAllString(content, "")
		}
		content = collapseBlankLines(content)
		if content != before && r.RuleID != "" {
			applied = append(applied, optimizer.AppliedRule{
				ID:        r.RuleID,
				SourceURL: r.Source,
			})
		}
		cp.Sections[i].Content = strings.TrimSpace(content)
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "RemoveBoilerplate", AppliedRules: applied}, nil
}

// ReduceAbsoluteLanguage compresses excessive MUST/ONLY while preserving critical patterns.
type ReduceAbsoluteLanguage struct {
	PreservePatterns []string
	RuleID           string
	Source           string
}

func (r ReduceAbsoluteLanguage) Name() string { return "RemoveBoilerplate" }

func (r ReduceAbsoluteLanguage) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	mustRe := regexp.MustCompile(`(?i)\b(MUST|ONLY|ALWAYS)\b`)
	var applied []optimizer.AppliedRule
	for i := range cp.Sections {
		if cp.Sections[i].Type == prompt.SectionTypeTask {
			continue
		}
		before := cp.Sections[i].Content
		content := mustRe.ReplaceAllStringFunc(before, func(m string) string {
			for _, preserve := range r.PreservePatterns {
				if strings.Contains(strings.ToLower(before), strings.ToLower(preserve)) {
					return m
				}
			}
			return strings.ToLower(m)
		})
		if content != before && r.RuleID != "" {
			applied = append(applied, optimizer.AppliedRule{ID: r.RuleID, SourceURL: r.Source})
		}
		cp.Sections[i].Content = content
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "RemoveBoilerplate", AppliedRules: applied}, nil
}

// BudgetAllocate assigns per-section token ceilings.
type BudgetAllocate struct {
	Counter optimizer.TokenCounter
}

func (b BudgetAllocate) Name() string { return "BudgetAllocate" }

func (b BudgetAllocate) Apply(_ context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	tb := budget.NewTokenBudget(cfg.MaxTokens)
	for i := range cp.Sections {
		if cp.Sections[i].Metadata == nil {
			cp.Sections[i].Metadata = make(map[string]string)
		}
		alloc := tb.Allocations[cp.Sections[i].Type]
		if alloc == 0 {
			alloc = cfg.MaxTokens / 10
		}
		cp.Sections[i].Metadata["token_budget"] = itoa(alloc)
	}
	return cp, optimizer.StageResult{StageName: "BudgetAllocate"}, nil
}

// TruncateByPriority shortens sections to fit max tokens, never truncating Task.
type TruncateByPriority struct {
	Counter optimizer.TokenCounter
}

func (TruncateByPriority) Name() string { return "TruncateByPriority" }

func (t TruncateByPriority) Apply(ctx context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	if t.Counter == nil {
		return cp, optimizer.StageResult{StageName: "TruncateByPriority"}, nil
	}

	total, err := t.Counter.Count(cp.Assemble())
	if err != nil {
		return nil, optimizer.StageResult{}, err
	}
	if total <= cfg.MaxTokens {
		return cp, optimizer.StageResult{StageName: "TruncateByPriority"}, nil
	}

	var truncated []string
	// Sort sections by ascending priority for truncation candidates.
	type indexed struct {
		idx      int
		priority int
	}
	var order []indexed
	for i, s := range cp.Sections {
		if s.Type == prompt.SectionTypeTask {
			continue
		}
		order = append(order, indexed{idx: i, priority: s.PriorityValue()})
	}
	// Bubble sort by priority ascending (low priority truncated first).
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			if order[j].priority < order[i].priority {
				order[i], order[j] = order[j], order[i]
			}
		}
	}

	for _, item := range order {
		select {
		case <-ctx.Done():
			return nil, optimizer.StageResult{}, ctx.Err()
		default:
		}
		total, err = t.Counter.Count(cp.Assemble())
		if err != nil {
			return nil, optimizer.StageResult{}, err
		}
		if total <= cfg.MaxTokens {
			break
		}
		idx := item.idx
		content := cp.Sections[idx].Content
		if len(content) < 50 {
			cp.Sections[idx].Content = ""
			truncated = append(truncated, cp.Sections[idx].ID)
			continue
		}
		// Remove ~30% each pass.
		cut := max(len(content)*7/10, 20)
		cp.Sections[idx].Content = truncateText(content, cut) + "\n\n[...truncated]"
		truncated = append(truncated, cp.Sections[idx].ID)
	}

	// If still over budget (e.g. single Task after format wrap), trim Task from the end.
	for {
		total, err = t.Counter.Count(cp.Assemble())
		if err != nil {
			return nil, optimizer.StageResult{}, err
		}
		if total <= cfg.MaxTokens {
			break
		}
		trimmed := false
		for i := range cp.Sections {
			if cp.Sections[i].Type != prompt.SectionTypeTask {
				continue
			}
			content := cp.Sections[i].Content
			if len(content) < 80 {
				break
			}
			cut := max(len(content)*6/10, 40)
			cp.Sections[i].Content = truncateText(content, cut) + "\n\n[...truncated]"
			truncated = append(truncated, cp.Sections[i].ID)
			trimmed = true
			break
		}
		if !trimmed {
			break
		}
	}

	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "TruncateByPriority", TruncatedSections: truncated}, nil
}

// EnforceRulesBudget limits Rules sections by word/line count.
type EnforceRulesBudget struct {
	MaxWords int
	MaxLines int
	RuleID   string
	Source   string
}

func (e EnforceRulesBudget) Name() string { return "TruncateByPriority" }

func (e EnforceRulesBudget) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	var truncated []string
	var applied []optimizer.AppliedRule
	for i := range cp.Sections {
		if cp.Sections[i].Type != prompt.SectionTypeRules {
			continue
		}
		content := cp.Sections[i].Content
		lines := strings.Split(content, "\n")
		words := len(strings.Fields(content))
		if e.MaxLines > 0 && len(lines) > e.MaxLines {
			lines = lines[:e.MaxLines]
			content = strings.Join(lines, "\n")
			truncated = append(truncated, cp.Sections[i].ID)
		}
		if e.MaxWords > 0 && words > e.MaxWords {
			fields := strings.Fields(content)
			content = strings.Join(fields[:e.MaxWords], " ")
			truncated = append(truncated, cp.Sections[i].ID)
		}
		cp.Sections[i].Content = content
	}
	if len(truncated) > 0 && e.RuleID != "" {
		applied = append(applied, optimizer.AppliedRule{ID: e.RuleID, SourceURL: e.Source})
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{
		StageName:         "TruncateByPriority",
		TruncatedSections: truncated,
		AppliedRules:      applied,
	}, nil
}

func collapseBlankLines(s string) string {
	re := regexp.MustCompile(`\n{3,}`)
	return strings.TrimSpace(re.ReplaceAllString(s, "\n\n"))
}

func truncateText(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// RemoveAgentKnownFacts strips common agent boilerplate.
type RemoveAgentKnownFacts struct {
	RuleID string
	Source string
}

func (RemoveAgentKnownFacts) Name() string { return "RemoveBoilerplate" }

var knownFacts = []string{
	"npm install",
	"git commit",
	"git push",
	"package manager",
	"version control",
}

func (r RemoveAgentKnownFacts) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	var applied []optimizer.AppliedRule
	for i := range cp.Sections {
		if cp.Sections[i].Type != prompt.SectionTypeRules {
			continue
		}
		before := cp.Sections[i].Content
		lines := strings.Split(before, "\n")
		var kept []string
		for _, line := range lines {
			lower := strings.ToLower(line)
			skip := false
			for _, fact := range knownFacts {
				if strings.Contains(lower, fact) {
					skip = true
					break
				}
			}
			if !skip {
				kept = append(kept, line)
			}
		}
		content := strings.Join(kept, "\n")
		if content != before && r.RuleID != "" {
			applied = append(applied, optimizer.AppliedRule{ID: r.RuleID, SourceURL: r.Source})
		}
		cp.Sections[i].Content = content
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "RemoveBoilerplate", AppliedRules: applied}, nil
}
