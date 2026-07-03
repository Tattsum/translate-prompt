package graph

import (
	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
)

// Resolver is the gqlgen root resolver.
type Resolver struct {
	OptimizeUC         *optimize.UseCase
	IntakeUC           *appintake.UseCase
	InvestigateEnabled bool
	LLMDefaults        budget.Config
}

// NewResolver wires application use cases into GraphQL resolvers.
func NewResolver(opt *optimize.UseCase, intake *appintake.UseCase, investigateEnabled bool, llmDefaults budget.Config) *Resolver {
	return &Resolver{
		OptimizeUC:         opt,
		IntakeUC:           intake,
		InvestigateEnabled: investigateEnabled,
		LLMDefaults:        llmDefaults,
	}
}

// EnrichConfig applies server-side defaults such as LLM settings.
func (r *Resolver) EnrichConfig(cfg budget.Config) budget.Config {
	return budget.EnrichLLMDefaults(cfg, r.LLMDefaults)
}
