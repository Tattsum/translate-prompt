package stages

import (
	"context"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
	infraast "github.com/Tattsum/translate-prompt/backend/infrastructure/ast"
)

// ASTParseStage parses markdown inside sections into an internal AST representation.
type ASTParseStage struct{}

func (ASTParseStage) Name() string { return "ASTParse" }

func (ASTParseStage) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	for i := range cp.Sections {
		if hasXMLStructure(cp.Sections[i].Content) {
			continue
		}
		_ = infraast.ParseMarkdown(cp.Sections[i].Content)
		if cp.Sections[i].Metadata == nil {
			cp.Sections[i].Metadata = make(map[string]string)
		}
		cp.Sections[i].Metadata["ast_parsed"] = "true"
	}
	return cp, optimizer.StageResult{StageName: "ASTParse"}, nil
}

// ASTCompressStage applies deterministic AST compression to section content.
type ASTCompressStage struct{}

func (ASTCompressStage) Name() string { return "ASTCompress" }

func (ASTCompressStage) Apply(_ context.Context, p *prompt.Prompt, _ budget.Config) (*prompt.Prompt, optimizer.StageResult, error) {
	cp := p.Clone()
	var applied []optimizer.AppliedRule
	for i := range cp.Sections {
		if hasXMLStructure(cp.Sections[i].Content) {
			continue
		}
		before := cp.Sections[i].Content
		doc := infraast.ParseMarkdown(before)
		compressed := infraast.Compress(doc)
		after := infraast.RenderMarkdown(compressed)
		if after != before {
			cp.Sections[i].Content = after
			applied = append(applied, optimizer.AppliedRule{
				ID:        "ast-compress",
				SourceURL: "https://github.com/yuin/goldmark",
				Method:    "ast",
			})
		}
	}
	cp.Raw = cp.Assemble()
	return cp, optimizer.StageResult{StageName: "ASTCompress", AppliedRules: applied}, nil
}

func hasXMLStructure(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "<") && strings.Contains(trimmed, ">")
}
