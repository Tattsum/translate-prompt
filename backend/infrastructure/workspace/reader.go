package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/intake"
	"github.com/Tattsum/translate-prompt/backend/domain/prompt"
)

const (
	maxFiles     = 20
	maxTotalSize = 100 * 1024
	maxDepth     = 2
	previewLen   = 200
)

var skipDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "dist": {}, "build": {},
}

var skipFiles = map[string]struct{}{
	".env": {}, ".env.local": {},
}

// BoundedFSReader scans a workspace with security limits.
type BoundedFSReader struct {
	Root string
}

// Investigate reads allowed files and suggests commands.
func (r *BoundedFSReader) Investigate() (intake.InvestigationResult, error) {
	root, err := filepath.Abs(r.Root)
	if err != nil {
		return intake.InvestigationResult{}, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return intake.InvestigationResult{}, fmt.Errorf("stat root: %w", err)
	}
	if !info.IsDir() {
		return intake.InvestigationResult{}, fmt.Errorf("workspace path is not a directory")
	}

	result := intake.InvestigationResult{}
	var totalSize int
	fileCount := 0

	targets := []string{
		"README.md", "CONTEXT.md", "go.mod", "package.json", "AGENTS.md",
		"Makefile", "Taskfile.yml",
	}

	for _, name := range targets {
		if fileCount >= maxFiles {
			break
		}
		path := filepath.Join(root, name)
		f, size, err := r.readFileSafe(root, path)
		if err != nil || f == nil {
			continue
		}
		totalSize += size
		fileCount++
		result.Files = append(result.Files, *f)
	}

	// .cursor/rules and .cursor/skills
	for _, sub := range []string{".cursor/rules", ".cursor/skills"} {
		dir := filepath.Join(root, sub)
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || fileCount >= maxFiles || totalSize >= maxTotalSize {
				return fs.SkipAll
			}
			if d.IsDir() {
				return nil
			}
			f, size, err := r.readFileSafe(root, path)
			if err != nil || f == nil {
				return nil
			}
			totalSize += size
			fileCount++
			result.Files = append(result.Files, *f)
			return nil
		})
	}

	// Directory listing depth 2
	entries, _ := os.ReadDir(root)
	var listing []string
	for _, e := range entries {
		if shouldSkipName(e.Name(), e.IsDir()) {
			continue
		}
		listing = append(listing, e.Name())
		if e.IsDir() {
			sub, _ := os.ReadDir(filepath.Join(root, e.Name()))
			for _, se := range sub {
				if !shouldSkipName(se.Name(), se.IsDir()) {
					listing = append(listing, e.Name()+"/"+se.Name())
				}
			}
		}
	}
	if len(listing) > 0 && fileCount < maxFiles {
		content := "Directory structure (depth 2):\n" + strings.Join(listing, "\n")
		result.Files = append(result.Files, intake.InvestigationFile{
			Path:           ".",
			SectionType:    prompt.SectionTypeCode,
			ContentPreview: truncatePreview(content),
			Content:        content,
		})
	}

	result.SuggestedCommands = detectCommands(root)
	return result, nil
}

func (r *BoundedFSReader) readFileSafe(root, path string) (*intake.InvestigationFile, int, error) {
	clean := filepath.Clean(path)
	if !strings.HasPrefix(clean, root) {
		return nil, 0, fmt.Errorf("path outside root")
	}

	// No symlink following: use Lstat
	li, err := os.Lstat(clean)
	if err != nil {
		return nil, 0, err
	}
	if li.Mode()&os.ModeSymlink != 0 {
		return nil, 0, fmt.Errorf("symlink skipped")
	}
	if li.IsDir() {
		return nil, 0, nil
	}
	if li.Size() > int64(maxTotalSize) {
		return nil, 0, fmt.Errorf("file too large")
	}

	base := filepath.Base(clean)
	if _, skip := skipFiles[base]; skip {
		return nil, 0, fmt.Errorf("skipped sensitive file")
	}
	if strings.HasSuffix(base, ".pem") {
		return nil, 0, fmt.Errorf("skipped pem file")
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, 0, err
	}
	content := string(data)
	rel, _ := filepath.Rel(root, clean)

	return &intake.InvestigationFile{
		Path:           rel,
		SectionType:    classifyFile(rel),
		ContentPreview: truncatePreview(content),
		Content:        content,
	}, len(data), nil
}

func classifyFile(path string) prompt.SectionType {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "skill"):
		return prompt.SectionTypeSkills
	case strings.Contains(lower, "rule"), strings.HasSuffix(lower, ".mdc"):
		return prompt.SectionTypeRules
	case strings.HasSuffix(lower, ".go"), strings.HasSuffix(lower, ".ts"), strings.HasSuffix(lower, ".tsx"):
		return prompt.SectionTypeCode
	default:
		return prompt.SectionTypeOther
	}
}

func detectCommands(root string) []string {
	var cmds []string
	checks := []struct {
		file string
		cmd  string
	}{
		{"Makefile", "make test"},
		{"Makefile", "make lint"},
		{"package.json", "npm test"},
		{"go.mod", "go test ./..."},
	}
	seen := make(map[string]struct{})
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			if _, ok := seen[c.cmd]; !ok {
				seen[c.cmd] = struct{}{}
				cmds = append(cmds, c.cmd)
			}
		}
	}
	return cmds
}

func shouldSkipName(name string, isDir bool) bool {
	if strings.HasPrefix(name, ".") && name != ".cursor" {
		return true
	}
	if isDir {
		if _, ok := skipDirs[name]; ok {
			return true
		}
	}
	return false
}

func truncatePreview(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= previewLen {
		return s
	}
	return s[:previewLen] + "..."
}
