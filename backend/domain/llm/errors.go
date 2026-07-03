package llm

import "errors"

// ErrBudgetExceeded is returned when LLM call or token limits are exceeded.
var ErrBudgetExceeded = errors.New("llm budget exceeded")

// ErrRefusal is returned when the model refuses the request.
var ErrRefusal = errors.New("llm refusal")

// ErrProviderUnavailable is returned when the provider is not configured or unreachable.
var ErrProviderUnavailable = errors.New("llm provider unavailable")

// ErrUnknownProvider is returned when routing selects an unsupported provider.
var ErrUnknownProvider = errors.New("llm unknown provider")
