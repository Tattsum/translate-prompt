package llm

import (
	"encoding/json"
	"strings"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

type intakeFindingsResponse struct {
	Findings []domainllm.ContextFinding `json:"findings"`
}

// ParseIntakeFindings extracts structured findings from an LLM intake response body.
func ParseIntakeFindings(content string) []domainllm.ContextFinding {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	var payload intakeFindingsResponse
	if err := json.Unmarshal([]byte(content), &payload); err == nil && len(payload.Findings) > 0 {
		return normalizeIntakeFindings(payload.Findings)
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), &payload); err == nil && len(payload.Findings) > 0 {
			return normalizeIntakeFindings(payload.Findings)
		}
	}

	return nil
}

func normalizeIntakeFindings(findings []domainllm.ContextFinding) []domainllm.ContextFinding {
	out := make([]domainllm.ContextFinding, 0, len(findings))
	for _, f := range findings {
		if strings.TrimSpace(f.Summary) == "" {
			continue
		}
		if f.Source == "" {
			f.Source = "llm"
		}
		out = append(out, f)
	}
	return out
}
