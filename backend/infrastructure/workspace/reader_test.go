package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Tattsum/translate-prompt/backend/infrastructure/workspace"
)

func TestBoundedFSReader_Investigate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# Test Project\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "Makefile"), []byte("test:\n\tgo test ./...\n"), 0o644)

	reader := &workspace.BoundedFSReader{Root: root}
	result, err := reader.Investigate()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) == 0 {
		t.Fatal("expected files")
	}
	foundREADME := false
	for _, f := range result.Files {
		if f.Path == "README.md" {
			foundREADME = true
		}
	}
	if !foundREADME {
		t.Fatal("README.md not found")
	}
}
