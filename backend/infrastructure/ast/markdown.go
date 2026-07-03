package ast

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"

	domainast "github.com/Tattsum/translate-prompt/backend/domain/ast"
)

var parser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// ParseMarkdown converts markdown into a domain AST document.
func ParseMarkdown(src string) *domainast.Document {
	src = strings.TrimSpace(src)
	if src == "" {
		return &domainast.Document{}
	}
	reader := text.NewReader([]byte(src))
	gmDoc := parser.Parser().Parse(reader)
	return convertDocument(gmDoc, reader)
}

func convertDocument(gmDoc ast.Node, reader text.Reader) *domainast.Document {
	doc := &domainast.Document{}
	for child := gmDoc.FirstChild(); child != nil; child = child.NextSibling() {
		if node := convertNode(child, reader); node != nil {
			doc.Children = append(doc.Children, node)
		}
	}
	return doc
}

func convertNode(n ast.Node, reader text.Reader) domainast.Node {
	switch n := n.(type) {
	case *ast.Paragraph:
		return domainast.Paragraph{Text: string(textContent(n, reader))}
	case *ast.Heading:
		return domainast.Heading{Level: n.Level, Text: string(textContent(n, reader))}
	case *ast.FencedCodeBlock:
		lang := string(n.Language(reader.Source()))
		return domainast.CodeBlock{Language: lang, Content: string(n.Lines().Value(reader.Source()))}
	case *ast.List:
		list := domainast.List{Ordered: n.IsOrdered()}
		for item := n.FirstChild(); item != nil; item = item.NextSibling() {
			li, ok := item.(*ast.ListItem)
			if !ok {
				continue
			}
			listItem := domainast.ListItem{Text: listItemText(li, reader)}
			list.Items = append(list.Items, listItem)
		}
		return list
	case *ast.TextBlock:
		return domainast.Paragraph{Text: string(textContent(n, reader))}
	default:
		return nil
	}
}

func listItemText(li *ast.ListItem, reader text.Reader) string {
	var parts []string
	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Paragraph:
			parts = append(parts, string(textContent(c, reader)))
		case *ast.TextBlock:
			parts = append(parts, string(textContent(c, reader)))
		}
	}
	return strings.Join(parts, " ")
}

func textContent(n ast.Node, reader text.Reader) []byte {
	var buf strings.Builder
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			buf.Write(t.Segment.Value(reader.Source()))
		}
	}
	return []byte(buf.String())
}

// RenderMarkdown serializes a domain AST back to markdown.
func RenderMarkdown(doc *domainast.Document) string {
	if doc == nil || len(doc.Children) == 0 {
		return ""
	}
	var b strings.Builder
	for i, node := range doc.Children {
		if i > 0 {
			switch doc.Children[i-1].(type) {
			case domainast.List:
				b.WriteString("\n")
			default:
				b.WriteString("\n\n")
			}
		}
		switch n := node.(type) {
		case domainast.Paragraph:
			b.WriteString(n.Text)
		case domainast.Heading:
			b.WriteString(strings.Repeat("#", n.Level))
			b.WriteString(" ")
			b.WriteString(n.Text)
		case domainast.CodeBlock:
			b.WriteString("```")
			b.WriteString(n.Language)
			b.WriteString("\n")
			b.WriteString(strings.TrimRight(n.Content, "\n"))
			b.WriteString("\n```")
		case domainast.List:
			for j, item := range n.Items {
				prefix := "- "
				if n.Ordered {
					prefix = itoa(j+1) + ". "
				}
				b.WriteString(prefix)
				b.WriteString(item.Text)
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// Compress applies deterministic AST transforms.
func Compress(doc *domainast.Document) *domainast.Document {
	if doc == nil {
		return &domainast.Document{}
	}
	out := &domainast.Document{}
	seenList := make(map[string]struct{})
	for _, node := range doc.Children {
		switch n := node.(type) {
		case domainast.List:
			n = dedupeList(n, seenList)
			if len(n.Items) > 0 {
				out.Children = append(out.Children, n)
			}
		case domainast.CodeBlock:
			n.Content = stripCodeComments(n.Content, n.Language)
			out.Children = append(out.Children, n)
		case domainast.Heading:
			if n.Level > 3 {
				n.Level = 3
			}
			out.Children = append(out.Children, n)
		default:
			out.Children = append(out.Children, node)
		}
	}
	return out
}

func dedupeList(list domainast.List, seen map[string]struct{}) domainast.List {
	var items []domainast.ListItem
	for _, item := range list.Items {
		key := strings.TrimSpace(strings.ToLower(item.Text))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, item)
	}
	list.Items = items
	return list
}

func stripCodeComments(content, lang string) string {
	lines := strings.Split(content, "\n")
	commentPrefix := commentPrefixFor(lang)
	if commentPrefix == "" {
		return content
	}
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, commentPrefix) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func commentPrefixFor(lang string) string {
	switch strings.ToLower(lang) {
	case "go", "rust", "java", "python", "py", "javascript", "js", "typescript", "ts", "c", "cpp":
		return "//"
	case "ruby", "rb", "sh", "bash", "yaml", "yml":
		return "#"
	default:
		return ""
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
