package bestpractice

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Tattsum/translate-prompt/backend/domain/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"gopkg.in/yaml.v3"
)

//go:embed rules/*.yaml
var rulesFS embed.FS

// Loader reads embedded best-practice YAML definitions.
type Loader struct {
	docs map[budget.TargetProfile]bestpractice.ProfileDocument
}

// NewLoader loads all profile YAML files from the embedded filesystem.
func NewLoader() (*Loader, error) {
	l := &Loader{docs: make(map[budget.TargetProfile]bestpractice.ProfileDocument)}
	err := fs.WalkDir(rulesFS, "rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := rulesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		var doc bestpractice.ProfileDocument
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		name := budget.TargetProfile(doc.Profile)
		l.docs[name] = doc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

// Load returns the TargetProfile bundle for a profile name.
func (l *Loader) Load(profile budget.TargetProfile) (bestpractice.TargetProfile, error) {
	common, ok := l.docs["common"]
	if !ok {
		return bestpractice.TargetProfile{}, fmt.Errorf("common profile not found")
	}
	doc, ok := l.docs[profile]
	if !ok {
		return bestpractice.TargetProfile{}, fmt.Errorf("profile %q not found", profile)
	}
	return bestpractice.TargetProfile{
		Name:    profile,
		Common:  common,
		Profile: doc,
	}, nil
}

// ProfileNames returns loaded profile names excluding common.
func (l *Loader) ProfileNames() []budget.TargetProfile {
	var names []budget.TargetProfile
	for name := range l.docs {
		if name == "common" {
			continue
		}
		names = append(names, name)
	}
	return names
}

// SourcePath returns the logical source path for a profile file.
func SourcePath(profile budget.TargetProfile) string {
	return filepath.Join("docs", "best-practices", string(profile)+".yaml")
}
