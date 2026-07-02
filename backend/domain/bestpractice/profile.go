package bestpractice

import "github.com/Tattsum/translate-prompt/backend/domain/budget"

// Rule defines a single best-practice automation rule from YAML.
type Rule struct {
	ID               string
	Description      string
	SourceURL        string `yaml:"source_url"`
	SourceSection    string `yaml:"source_section"`
	Automatable      bool
	Pipeline         string // format | compress
	Stage            string
	Action           string
	Patterns         []string
	PreservePatterns []string `yaml:"preserve_patterns"`
	IntakeOnFailure  bool     `yaml:"intake_on_failure"`
	Limits           map[string]int
	Condition        map[string]string
}

// ProfileDocument is the YAML schema for a target profile.
type ProfileDocument struct {
	Profile               string `yaml:"profile"`
	Version               string `yaml:"version"`
	LastReviewed          string `yaml:"last_reviewed"`
	OutputTemplate        string `yaml:"output_template"`
	OutputTemplateChat    string `yaml:"output_template_chat"`
	OutputTemplateMDC     string `yaml:"output_template_mdc"`
	OutputTemplateSession string `yaml:"output_template_session_brief"`
	References            []Reference
	Rules                 []Rule
}

// Reference links to official documentation.
type Reference struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
}

// TargetProfile bundles common + profile-specific rules.
type TargetProfile struct {
	Name    budget.TargetProfile
	Common  ProfileDocument
	Profile ProfileDocument
}

// AllRules returns common rules followed by profile-specific rules.
func (tp *TargetProfile) AllRules() []Rule {
	out := make([]Rule, 0, len(tp.Common.Rules)+len(tp.Profile.Rules))
	out = append(out, tp.Common.Rules...)
	out = append(out, tp.Profile.Rules...)
	return out
}

// RulesForPipeline returns rules matching the pipeline name.
func (tp *TargetProfile) RulesForPipeline(pipeline string) []Rule {
	var out []Rule
	for _, r := range tp.AllRules() {
		if r.Pipeline == pipeline && r.Automatable {
			out = append(out, r)
		}
	}
	return out
}

// RulesForStage returns rules for a specific stage name.
func (tp *TargetProfile) RulesForStage(stage string) []Rule {
	var out []Rule
	for _, r := range tp.AllRules() {
		if r.Stage == stage && r.Automatable {
			out = append(out, r)
		}
	}
	return out
}
