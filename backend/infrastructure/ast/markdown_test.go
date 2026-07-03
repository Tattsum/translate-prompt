package ast_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	infraast "github.com/Tattsum/translate-prompt/backend/infrastructure/ast"
)

func TestCompress_ListDedupeGolden(t *testing.T) {
	t.Parallel()

	input := readTestdata(t, "list_dedupe.input.md")
	want := strings.TrimSpace(readTestdata(t, "list_dedupe.golden.md"))

	doc := infraast.ParseMarkdown(input)
	compressed := infraast.Compress(doc)
	got := infraast.RenderMarkdown(compressed)

	if strings.TrimSpace(got) != want {
		t.Fatalf("golden mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
