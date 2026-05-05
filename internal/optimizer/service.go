package optimizer

import (
	"context"

	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/domain"
)

type TemporalHistoryClient interface {
	Config() config.TemporalConfig
}

type Service struct {
	temporal TemporalHistoryClient
}

func NewService(temporal TemporalHistoryClient) *Service {
	return &Service{temporal: temporal}
}

func (s *Service) AnalyzeWorkflow(context.Context, string) (domain.WorkflowAnalysis, error) {
	return domain.WorkflowAnalysis{}, domain.ErrNotImplemented
}

var _ domain.Optimizer = (*Service)(nil)
