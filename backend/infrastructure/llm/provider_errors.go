package llm

import (
	"encoding/json"
	"errors"
	"strings"

	domainllm "github.com/Tattsum/translate-prompt/backend/domain/llm"
)

type apiErrorEnvelope struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func mapProviderError(provider string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domainllm.ErrRefusal) {
		return err
	}
	if isRefusalFromAPIError(err) {
		return domainllm.ErrRefusal
	}
	return errors.Join(errors.New(provider), err)
}

func isRefusalFromAPIError(err error) bool {
	for e := err; e != nil; e = errors.Unwrap(e) {
		msg := e.Error()
		start := strings.Index(msg, "{")
		end := strings.LastIndex(msg, "}")
		if start < 0 || end <= start {
			continue
		}
		var envelope apiErrorEnvelope
		if json.Unmarshal([]byte(msg[start:end+1]), &envelope) != nil {
			continue
		}
		typ := strings.ToLower(envelope.Error.Type)
		body := strings.ToLower(envelope.Error.Message)
		if strings.Contains(typ, "refus") || strings.Contains(body, "refus") {
			return true
		}
	}
	return false
}
