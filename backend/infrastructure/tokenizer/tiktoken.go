package tokenizer

import (
	"fmt"
	"sync"

	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
	"github.com/pkoukk/tiktoken-go"
)

// Tiktoken implements optimizer.TokenCounter using tiktoken-go.
type Tiktoken struct {
	encoding *tiktoken.Tiktoken
	mu       sync.Mutex
}

// New creates a tokenizer for the given encoding name.
func New(encodingName string) (*Tiktoken, error) {
	if encodingName == "" {
		encodingName = "cl100k_base"
	}
	enc, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, fmt.Errorf("get encoding %q: %w", encodingName, err)
	}
	return &Tiktoken{encoding: enc}, nil
}

// Count returns the token count for text.
func (t *Tiktoken) Count(text string) (int, error) {
	if t == nil || t.encoding == nil {
		return 0, fmt.Errorf("tokenizer not initialized")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	tokens := t.encoding.Encode(text, nil, nil)
	return len(tokens), nil
}

var _ optimizer.TokenCounter = (*Tiktoken)(nil)
