package bestpractice

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	domainbp "github.com/Tattsum/translate-prompt/backend/domain/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

// ApplyBestPracticeProfile applies profile-specific rule actions.
type ApplyBestPracticeProfile struct {
	Loader *Loader
}

func (a ApplyBestPracticeProfile) Name() string { return "ApplyBestPracticeProfile" }

func (a ApplyBestPracticeProfile) Apply(_ context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	tp, err := a.Loader.Load(cfg.TargetProfile)
	if err != nil {
		return nil, optimizer.StageResult{}, err
	}

	var applied []optimizer.AppliedRule
	for _, rule := range tp.RulesForStage("ApplyBestPracticeProfile") {
		switch rule.Action {
		case "rewrite_imperative":
			applied = append(applied, applyRewriteImperative(cp, rule)...)
		case "append_verification_commands":
			applied = append(applied, applyVerificationCommands(cp, cfg, rule)...)
		case "split_rules_and_skills":
			applied = append(applied, applySplitRulesSkills(cp, rule)...)
		case "replace_code_with_at_reference":
			applied = append(applied, applyAtReference(cp, rule)...)
		case "make_actionable":
			applied = append(applied, applyMakeActionable(cp, rule)...)
		case "add_concrete_placeholders":
			applied = append(applied, applyConcretePlaceholders(cp, rule)...)
		case "detect_overscope":
			applied = append(applied, applyDetectOverscope(cp, rule)...)
		case "split_knowledge_playbook":
			applied = append(applied, applySplitKnowledge(cp, rule)...)
		case "suggest_agents_md":
			applied = append(applied, applySuggestAgentsMD(cp, rule)...)
		case "split_composite_rules":
			applied = append(applied, applySplitComposite(cp, rule)...)
		}
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "ApplyBestPracticeProfile", AppliedRules: applied}, nil
}

// ReorderOutcomeFirst moves Task sections to the front.
type ReorderOutcomeFirst struct {
	Loader *Loader
}

func (r ReorderOutcomeFirst) Name() string { return "ReorderOutcomeFirst" }

func (r ReorderOutcomeFirst) Apply(_ context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	tp, err := r.Loader.Load(cfg.TargetProfile)
	if err != nil {
		return nil, optimizer.StageResult{}, err
	}

	var applied []optimizer.AppliedRule
	for _, rule := range tp.RulesForStage("ReorderOutcomeFirst") {
		switch rule.Action {
		case "outcome_first_reorder":
			cp.Sections = reorderOutcomeFirst(cp.Sections)
			applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
		case "consolidate_references":
			applied = append(applied, consolidateReferences(cp, rule)...)
		}
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "ReorderOutcomeFirst", AppliedRules: applied}, nil
}

// WrapStructure applies output templates per profile.
type WrapStructure struct {
	Loader  *Loader
	Counter optimizer.TokenCounter
}

func (w WrapStructure) Name() string { return "WrapStructure" }

func (w WrapStructure) Apply(_ context.Context, p *prompt.Prompt, cfg budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	tp, err := w.Loader.Load(cfg.TargetProfile)
	if err != nil {
		return nil, optimizer.StageResult{}, err
	}

	var applied []optimizer.AppliedRule
	inputTokens := 0
	if w.Counter != nil {
		inputTokens, _ = w.Counter.Count(cp.Assemble())
	}

	for _, rule := range tp.RulesForStage("WrapStructure") {
		if !ruleMatchesCondition(rule, inputTokens, cfg) {
			continue
		}
		switch rule.Action {
		case "wrap_xml_tags":
			out, ok := wrapXML(cp, tp)
			if ok {
				cp = out
				applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
			}
		case "apply_four_part_template":
			out, ok := wrapFourPart(cp, tp.Profile.OutputTemplate, rule)
			if ok {
				cp = out
				applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
			}
		case "apply_three_part_template":
			out, ok := wrapThreePart(cp, tp.Profile.OutputTemplate, rule)
			if ok {
				cp = out
				applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
			}
		case "consolidate_examples":
			applied = append(applied, consolidateExamples(cp, rule)...)
		case "generate_mdc_artifact":
			applied = append(applied, generateMDCArtifact(cp, tp, rule)...)
		case "generate_scope_out":
			applied = append(applied, generateScopeOut(cp, rule)...)
		case "wrap_documents":
			if inputTokens > 20000 {
				applied = append(applied, wrapDocuments(cp, rule)...)
			}
		}
	}

	// Profile-specific final assembly for openai (same template as codex).
	if cfg.TargetProfile == budget.ProfileOpenAI && tp.Profile.OutputTemplate != "" {
		out, ok := wrapFourPart(cp, tp.Profile.OutputTemplate, domainbp.Rule{ID: "openai-template"})
		if ok {
			cp = out
		}
	}

	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "WrapStructure", AppliedRules: applied}, nil
}

// AssembleWithProfile sets final raw text from sections.
type AssembleWithProfile struct{}

func (AssembleWithProfile) Name() string { return "AssembleWithProfile" }

func (AssembleWithProfile) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "AssembleWithProfile"}, nil
}

func applyRewriteImperative(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	replacements := map[string]string{
		"してほしい":  "せよ",
		"改善して":   "改善する",
		"なんとかして": "具体的に実行する",
	}
	var applied []optimizer.AppliedRule
	for i := range p.Sections {
		if p.Sections[i].Type != prompt.SectionTypeTask {
			continue
		}
		content := p.Sections[i].Content
		changed := false
		for _, pat := range rule.Patterns {
			if strings.Contains(content, pat) {
				if rep, ok := replacements[pat]; ok {
					content = strings.ReplaceAll(content, pat, rep)
					changed = true
				}
			}
		}
		if changed {
			p.Sections[i].Content = content
			applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
		}
	}
	return applied
}

func applyVerificationCommands(p *prompt.Prompt, cfg budget.Config, rule domainbp.Rule) []optimizer.AppliedRule {
	if len(cfg.VerificationCommands) == 0 {
		return nil
	}
	cmds := strings.Join(cfg.VerificationCommands, ", ")
	var applied []optimizer.AppliedRule
	for i := range p.Sections {
		if p.Sections[i].Type == prompt.SectionTypeRules || p.Sections[i].Type == prompt.SectionTypeTask {
			if !strings.Contains(p.Sections[i].Content, cmds) {
				p.Sections[i].Content += "\n\nVerify: " + cmds
				applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
			}
		}
	}
	return applied
}

func applySplitRulesSkills(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	var applied []optimizer.AppliedRule
	for i := range p.Sections {
		if p.Sections[i].Type != prompt.SectionTypeRules {
			continue
		}
		lines := strings.Split(p.Sections[i].Content, "\n")
		if len(lines) > 5 {
			p.Sections[i].Type = prompt.SectionTypeSkills
			p.Sections[i].Content = "@skill-workflow\n" + p.Sections[i].Content
			applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
		}
	}
	return applied
}

func applyAtReference(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	codeRe := regexp.MustCompile("(?s)```[\\w]*\\n(.*?)```")
	var applied []optimizer.AppliedRule
	for i := range p.Sections {
		if p.Sections[i].Type != prompt.SectionTypeCode {
			continue
		}
		if codeRe.MatchString(p.Sections[i].Content) {
			p.Sections[i].Content = "@src/relevant-file.go"
			applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
		}
	}
	return applied
}

func applyMakeActionable(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	replacements := map[string]string{
		"きれいに":           "lint と formatter を通す",
		"clean code":     "既存の命名規約に従う",
		"best practices": "プロジェクトの docs/architecture.md に従う",
	}
	var applied []optimizer.AppliedRule
	for i := range p.Sections {
		content := p.Sections[i].Content
		changed := false
		for _, pat := range rule.Patterns {
			if strings.Contains(strings.ToLower(content), strings.ToLower(pat)) {
				if rep, ok := replacements[strings.ToLower(pat)]; ok {
					re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(pat))
					content = re.ReplaceAllString(content, rep)
					changed = true
				}
			}
		}
		if changed {
			p.Sections[i].Content = content
			applied = append(applied, optimizer.AppliedRule{ID: rule.ID, SourceURL: rule.SourceURL})
		}
	}
	return applied
}

func applyConcretePlaceholders(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	for i := range p.Sections {
		if p.Sections[i].Type == prompt.SectionTypeTask {
			if !strings.Contains(p.Sections[i].Content, "repo") {
				p.Sections[i].Content += "\n\nRepo: <owner/repo>\nBranch: main\nFiles: <path/to/file>"
			}
		}
	}
	return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
}

func applyDetectOverscope(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	task := p.ContentByType(prompt.SectionTypeTask)
	if strings.Count(task, " and ") >= 3 || strings.Count(task, "、") >= 5 {
		p.Artifacts["overscope_warning"] = "Consider splitting into multiple PR-sized tasks"
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func applySplitKnowledge(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	for i := range p.Sections {
		if p.Sections[i].Type == prompt.SectionTypeRules && strings.Contains(p.Sections[i].Content, "How") {
			p.Sections = append(p.Sections, prompt.Section{
				ID: "playbook-0", Type: prompt.SectionTypeSkills, Content: p.Sections[i].Content,
			})
			return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
		}
	}
	return nil
}

func applySuggestAgentsMD(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	totalLines := 0
	for _, s := range p.SectionsByType(prompt.SectionTypeRules) {
		totalLines += strings.Count(s.Content, "\n") + 1
	}
	if totalLines < 30 {
		p.Artifacts["suggest_agents_md"] = true
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func applySplitComposite(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	var suggestions []map[string]string
	for _, s := range p.SectionsByType(prompt.SectionTypeRules) {
		if strings.Count(s.Content, "\n\n") >= 3 {
			suggestions = append(suggestions, map[string]string{
				"source": s.ID,
				"hint":   "Split into focused .mdc files by topic",
			})
		}
	}
	if len(suggestions) > 0 {
		p.Artifacts["rule_split_suggestions"] = suggestions
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func consolidateReferences(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	urlRe := regexp.MustCompile(`https?://\S+`)
	var refs []string
	for i := range p.Sections {
		matches := urlRe.FindAllString(p.Sections[i].Content, -1)
		if len(matches) == 0 {
			continue
		}
		for _, m := range matches {
			p.Sections[i].Content = strings.ReplaceAll(p.Sections[i].Content, m, "")
			refs = append(refs, m)
		}
	}
	if len(refs) > 0 {
		p.Sections = append(p.Sections, prompt.Section{
			ID: "refs-0", Type: prompt.SectionTypeHistory, Content: "References:\n" + strings.Join(refs, "\n"),
		})
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func consolidateExamples(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	var examples []string
	var kept []prompt.Section
	for _, s := range p.Sections {
		if s.Type == prompt.SectionTypeCode {
			examples = append(examples, s.Content)
			continue
		}
		kept = append(kept, s)
	}
	if len(examples) > 0 {
		kept = append(kept, prompt.Section{
			ID: "examples-0", Type: prompt.SectionTypeCode, Content: strings.Join(examples, "\n\n---\n\n"),
		})
		p.Sections = kept
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func generateMDCArtifact(p *prompt.Prompt, tp domainbp.TargetProfile, rule domainbp.Rule) []optimizer.AppliedRule {
	var mdc []map[string]string
	for _, s := range p.SectionsByType(prompt.SectionTypeRules) {
		if len(s.Content) < 20 {
			continue
		}
		body := strings.TrimSpace(s.Content)
		tmpl := tp.Profile.OutputTemplateMDC
		if tmpl == "" {
			tmpl = "---\ndescription: {{description}}\n---\n\n{{rule_body}}"
		}
		content := strings.ReplaceAll(tmpl, "{{description}}", "Project rule")
		content = strings.ReplaceAll(content, "{{globs}}", "**/*")
		content = strings.ReplaceAll(content, "{{always_apply}}", "false")
		content = strings.ReplaceAll(content, "{{rule_body}}", body)
		mdc = append(mdc, map[string]string{
			"filename": s.ID + ".mdc",
			"content":  content,
		})
	}
	if len(mdc) > 0 {
		p.Artifacts["cursor_mdc_suggestions"] = mdc
		return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
	}
	return nil
}

func generateScopeOut(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	p.Artifacts["scope_out"] = "Generated configs, unrelated modules, production data"
	return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
}

func wrapDocuments(p *prompt.Prompt, rule domainbp.Rule) []optimizer.AppliedRule {
	for i := range p.Sections {
		if p.Sections[i].Type == prompt.SectionTypeHistory {
			p.Sections[i].Content = fmt.Sprintf("<document index=\"%d\">\n%s\n</document>", i, p.Sections[i].Content)
		}
	}
	return []optimizer.AppliedRule{{ID: rule.ID, SourceURL: rule.SourceURL}}
}

func wrapXML(p *prompt.Prompt, tp domainbp.TargetProfile) (*prompt.Prompt, bool) {
	tmpl := tp.Profile.OutputTemplate
	if tmpl == "" {
		return p, false
	}
	out := strings.ReplaceAll(tmpl, "{{task}}", p.ContentByType(prompt.SectionTypeTask))
	out = strings.ReplaceAll(out, "{{context}}", p.ContentByType(prompt.SectionTypeHistory))
	out = strings.ReplaceAll(out, "{{rules}}", p.ContentByType(prompt.SectionTypeRules))
	out = strings.ReplaceAll(out, "{{examples}}", p.ContentByType(prompt.SectionTypeCode))
	out = strings.ReplaceAll(out, "{{constraints}}", extractConstraints(p))
	cp := p.Clone()
	cp.Sections = []prompt.Section{{ID: "wrapped-0", Type: prompt.SectionTypeTask, Content: strings.TrimSpace(out)}}
	return cp, true
}

func wrapFourPart(p *prompt.Prompt, tmpl string, rule domainbp.Rule) (*prompt.Prompt, bool) {
	if tmpl == "" {
		return p, false
	}
	goal := p.ContentByType(prompt.SectionTypeTask)
	context := p.ContentByType(prompt.SectionTypeHistory)
	if context == "" {
		context = p.ContentByType(prompt.SectionTypeOther)
	}
	constraints := extractConstraints(p)
	doneWhen := extractDoneWhen(p)

	out := strings.ReplaceAll(tmpl, "{{goal}}", goal)
	out = strings.ReplaceAll(out, "{{context}}", context)
	out = strings.ReplaceAll(out, "{{constraints}}", constraints)
	out = strings.ReplaceAll(out, "{{done_when}}", doneWhen)

	cp := p.Clone()
	cp.Sections = []prompt.Section{{ID: "wrapped-0", Type: prompt.SectionTypeTask, Content: strings.TrimSpace(out)}}
	return cp, true
}

func wrapThreePart(p *prompt.Prompt, tmpl string, rule domainbp.Rule) (*prompt.Prompt, bool) {
	if tmpl == "" {
		return p, false
	}
	what := p.ContentByType(prompt.SectionTypeTask)
	how := p.ContentByType(prompt.SectionTypeSkills)
	if how == "" {
		how = p.ContentByType(prompt.SectionTypeCode)
	}
	result := extractDoneWhen(p)

	out := strings.ReplaceAll(tmpl, "{{what}}", what)
	out = strings.ReplaceAll(out, "{{how}}", how)
	out = strings.ReplaceAll(out, "{{result}}", result)

	cp := p.Clone()
	cp.Sections = []prompt.Section{{ID: "wrapped-0", Type: prompt.SectionTypeTask, Content: strings.TrimSpace(out)}}
	return cp, true
}

func extractConstraints(p *prompt.Prompt) string {
	var parts []string
	for _, s := range p.SectionsByType(prompt.SectionTypeRules) {
		if strings.Contains(strings.ToLower(s.Content), "constraint") ||
			strings.Contains(strings.ToLower(s.Content), "must") ||
			strings.Contains(strings.ToLower(s.Content), "禁止") {
			parts = append(parts, s.Content)
		}
	}
	if len(parts) == 0 {
		return p.ContentByType(prompt.SectionTypeRules)
	}
	return strings.Join(parts, "\n")
}

func extractDoneWhen(p *prompt.Prompt) string {
	task := p.ContentByType(prompt.SectionTypeTask)
	if idx := strings.Index(strings.ToLower(task), "done"); idx >= 0 {
		return strings.TrimSpace(task[idx:])
	}
	if idx := strings.Index(task, "完了"); idx >= 0 {
		return strings.TrimSpace(task[idx:])
	}
	return "All tests pass and acceptance criteria are met."
}

func reorderOutcomeFirst(sections []prompt.Section) []prompt.Section {
	var tasks, rest []prompt.Section
	for _, s := range sections {
		if s.Type == prompt.SectionTypeTask {
			tasks = append(tasks, s)
		} else {
			rest = append(rest, s)
		}
	}
	return append(tasks, rest...)
}

func ruleMatchesCondition(rule domainbp.Rule, inputTokens int, cfg budget.Config) bool {
	for k, v := range rule.Condition {
		switch k {
		case "input_tokens_gt":
			threshold, err := strconv.Atoi(v)
			if err != nil || inputTokens <= threshold {
				return false
			}
		case "output_mode":
			if cfg.OutputMode != v {
				return false
			}
		case "total_rule_lines_lt":
			// handled in suggest_agents_md action
		}
	}
	return true
}
