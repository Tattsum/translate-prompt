package intake

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/llm"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/workspace"
)

// UseCase handles analyze, investigate, and merge flows.
type UseCase struct {
	Loader    *infraBP.Loader
	Completer llm.Completer
}

// NewUseCase creates an intake use case.
func NewUseCase(loader *infraBP.Loader) *UseCase {
	return &UseCase{Loader: loader}
}

// WithCompleter attaches an LLM completer for intake analysis.
func (uc *UseCase) WithCompleter(c llm.Completer) *UseCase {
	uc.Completer = c
	return uc
}

// Analyze detects ambiguities in the prompt.
func (uc *UseCase) Analyze(ctx context.Context, raw string, cfg budget.Config) (domainintake.AnalyzeResult, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return domainintake.AnalyzeResult{
			Status: domainintake.StatusNeedsInput,
			Questions: []domainintake.Question{{
				ID: "prompt", Text: "最適化するプロンプトを入力してください。",
			}},
		}, nil
	}

	var questions []domainintake.Question
	findings := heuristicFindings(text, cfg)

	var llmFindings []domainintake.Finding
	if cfg.DeepDive && cfg.LLMEnabled && uc.Completer != nil {
		llm.ResetBudgetIfSupported(uc.Completer)
		ctxFindings := make([]llm.ContextFinding, 0, len(findings))
		for _, f := range findings {
			ctxFindings = append(ctxFindings, domainintake.ToContextFinding(f))
		}
		intent := llm.CompletionIntent{
			Purpose:       llm.PurposeIntakeAnalyze,
			TargetProfile: cfg.TargetProfile,
			InputContent:  text,
			Context: llm.CompletionContext{
				HeuristicFindings: ctxFindings,
				PromptSections:    summarizeSections(prompt.New(text)),
			},
		}
		outcome, err := uc.Completer.Complete(ctx, intent, llm.CompletionBudgetFrom(cfg))
		if err != nil {
			if errors.Is(err, llm.ErrBudgetExceeded) {
				slog.WarnContext(ctx, "intake llm budget exceeded",
					"target_profile", cfg.TargetProfile,
					"purpose", llm.PurposeIntakeAnalyze,
				)
			} else {
				slog.WarnContext(ctx, "intake llm analyze failed",
					"target_profile", cfg.TargetProfile,
					"purpose", llm.PurposeIntakeAnalyze,
					"err", err,
				)
			}
		} else {
			for _, cf := range outcome.Findings {
				f := domainintake.FindingFromContext(cf)
				f.Source = domainintake.FindingSourceLLM
				llmFindings = append(llmFindings, f)
			}
		}
	}

	merged := domainintake.MergeFindings(findings, llmFindings)
	questions = domainintake.QuestionsFromFindings(merged)

	if cfg.DeepDive && len(questions) > 0 {
		return domainintake.AnalyzeResult{
			Status:    domainintake.StatusNeedsInput,
			Questions: questions,
			Findings:  merged,
		}, nil
	}

	return domainintake.AnalyzeResult{
		Status:   domainintake.StatusReady,
		Prompt:   text,
		Findings: findingsForResponse(cfg, merged),
	}, nil
}

// Investigate scans the workspace for context.
func (uc *UseCase) Investigate(_ context.Context, workspacePath string, _ budget.TargetProfile) (domainintake.InvestigationResult, error) {
	reader := &workspace.BoundedFSReader{Root: workspacePath}
	return reader.Investigate()
}

// MergeAnswers incorporates user answers into the prompt.
func (uc *UseCase) MergeAnswers(raw string, answers map[string]string) string {
	if len(answers) == 0 {
		return raw
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(raw))
	b.WriteString("\n\n## Clarifications\n")
	for id, ans := range answers {
		if strings.TrimSpace(ans) == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(id)
		b.WriteString(": ")
		b.WriteString(ans)
		b.WriteString("\n")
	}
	return b.String()
}

// MergeContext appends investigation files as sections.
func MergeContext(raw string, inv domainintake.InvestigationResult) string {
	p := prompt.New(raw)
	for i, f := range inv.Files {
		p.Sections = append(p.Sections, prompt.Section{
			ID:      fmt.Sprintf("ws-%d", i),
			Type:    f.SectionType,
			Content: "## " + f.Path + "\n\n" + f.Content,
		})
	}
	return p.Assemble()
}

func heuristicFindings(text string, cfg budget.Config) []domainintake.Finding {
	questions := detectAmbiguities(text, cfg)
	findings := make([]domainintake.Finding, 0, len(questions))
	for _, q := range questions {
		findings = append(findings, questionToFinding(q))
	}
	return findings
}

func questionToFinding(q domainintake.Question) domainintake.Finding {
	category := questionCategory(q.ID)
	return domainintake.Finding{
		ID:       q.ID,
		Category: category,
		Severity: 3,
		RuleID:   q.RuleID,
		Summary:  q.Text,
		Source:   domainintake.FindingSourceHeuristic,
	}
}

func questionCategory(id string) string {
	switch id {
	case "goal":
		return "goal_unclear"
	case "scope":
		return "scope_missing"
	case "priority":
		return "contradiction"
	case "acceptance":
		return "acceptance_missing"
	default:
		return id
	}
}

func summarizeSections(p *prompt.Prompt) []llm.PromptSectionSummary {
	out := make([]llm.PromptSectionSummary, 0, len(p.Sections))
	for i, s := range p.Sections {
		content := s.Content
		if len(content) > 2000 {
			content = content[:2000]
		}
		out = append(out, llm.PromptSectionSummary{
			Ref:     llm.SectionRef{Index: i, ID: s.ID, Type: s.Type},
			Content: content,
		})
	}
	return out
}

func detectAmbiguities(text string, cfg budget.Config) []domainintake.Question {
	var qs []domainintake.Question
	lower := strings.ToLower(text)

	if !hasGoal(text) {
		qs = append(qs, domainintake.Question{
			ID: "goal", Text: "成功条件・完了の定義は？", RuleID: profileRuleID(cfg.TargetProfile, "goal"),
		})
	}
	if !hasScope(text) {
		qs = append(qs, domainintake.Question{
			ID: "scope", Text: "対象ファイル・モジュールの範囲は？", RuleID: "scope",
		})
	}
	if hasContradiction(text) {
		qs = append(qs, domainintake.Question{
			ID: "priority", Text: "矛盾する要求があります。どちらを優先しますか？", RuleID: "contradiction",
		})
	}
	if !hasAcceptance(text) {
		qs = append(qs, domainintake.Question{
			ID: "acceptance", Text: "完了条件・検証コマンドは？", RuleID: "acceptance",
		})
	}
	if cfg.MaxTokens <= 0 {
		qs = append(qs, domainintake.Question{
			ID: "budget", Text: "max-tokens・TargetProfile は？", RuleID: "budget",
		})
	}

	switch cfg.TargetProfile {
	case budget.ProfileDevin:
		if strings.Count(text, " and ") >= 3 {
			qs = append(qs, domainintake.Question{
				ID: "devin-split", Text: "1 セッション 1 PR に分割しますか？", RuleID: "devin-scope-split",
			})
		}
		if !strings.Contains(lower, "scope out") && !strings.Contains(text, "触らない") {
			qs = append(qs, domainintake.Question{
				ID: "devin-scope-out", Text: "触ってはいけない領域は？", RuleID: "devin-scope-out",
			})
		}
	case budget.ProfileCursor:
		if strings.Count(text, "always") > 2 {
			qs = append(qs, domainintake.Question{
				ID: "cursor-always", Text: "どの規約を常時適用に残すか？", RuleID: "cursor-rules-budget",
			})
		}
		if strings.Contains(lower, "skill") && strings.Contains(lower, "rule") {
			qs = append(qs, domainintake.Question{
				ID: "cursor-rules-skills", Text: "手順は Rule と Skill のどちら？", RuleID: "cursor-rules-skills-split",
			})
		}
	}

	return qs
}

func hasGoal(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"goal", "目的", "完了", "done when", "成功", "what"}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func hasScope(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"scope", "範囲", "ファイル", "module", "package", "dir/"}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func hasContradiction(text string) bool {
	lower := strings.ToLower(text)
	pairs := [][2]string{
		{"fast", "thorough"},
		{"速く", "丁寧"},
		{"minimal", "comprehensive"},
	}
	for _, p := range pairs {
		if strings.Contains(lower, p[0]) && strings.Contains(lower, p[1]) {
			return true
		}
	}
	return false
}

func hasAcceptance(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"test", "lint", "acceptance", "verify", "検証", "make test"}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func profileRuleID(profile budget.TargetProfile, kind string) string {
	switch profile {
	case budget.ProfileCodex:
		return "codex-four-part"
	case budget.ProfileClaude:
		return "claude-explicit"
	case budget.ProfileOpenAI:
		return "openai-outcome-first"
	case budget.ProfileDevin:
		return "devin-three-part"
	case budget.ProfileCursor:
		return "cursor-actionable"
	default:
		return kind
	}
}

func findingsForResponse(cfg budget.Config, merged []domainintake.Finding) []domainintake.Finding {
	if !cfg.DeepDive {
		return nil
	}
	return merged
}
