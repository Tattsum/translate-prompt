package ast

// Node is a lightweight AST node for deterministic compression.
type Node interface {
	node()
}

// Document is the root of a parsed section content tree.
type Document struct {
	Children []Node
}

// Paragraph is plain text content.
type Paragraph struct {
	Text string
}

func (Paragraph) node() {}

// Heading is a section heading.
type Heading struct {
	Level int
	Text  string
}

func (Heading) node() {}

// List is an ordered or unordered list.
type List struct {
	Ordered bool
	Items   []ListItem
}

func (List) node() {}

// ListItem is a single list entry.
type ListItem struct {
	Text     string
	Children []Node
}

// CodeBlock is a fenced code block.
type CodeBlock struct {
	Language string
	Content  string
}

func (CodeBlock) node() {}
