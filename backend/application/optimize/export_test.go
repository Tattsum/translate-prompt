package optimize

import (
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/domain/optimizer"
)

// ExportBuildCompressPipeline exposes the compress pipeline for integration tests.
func (uc *UseCase) ExportBuildCompressPipeline(cfg budget.Config) (*optimizer.Pipeline, error) {
	return uc.buildCompressPipeline(cfg)
}
