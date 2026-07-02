package graph

import (
	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
)

// Resolver is the gqlgen root resolver.
type Resolver struct {
	OptimizeUC *optimize.UseCase
	IntakeUC   *appintake.UseCase
}

// NewResolver wires application use cases into GraphQL resolvers.
func NewResolver(opt *optimize.UseCase, intake *appintake.UseCase) *Resolver {
	return &Resolver{OptimizeUC: opt, IntakeUC: intake}
}
