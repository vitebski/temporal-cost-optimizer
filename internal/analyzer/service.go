package analyzer

import (
	"context"

	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/domain"
)

type TemporalUsageClient interface {
	Config() config.TemporalConfig
}

type Service struct {
	temporal TemporalUsageClient
}

func NewService(temporal TemporalUsageClient) *Service {
	return &Service{temporal: temporal}
}

func (s *Service) TopNamespaces(context.Context, int) ([]domain.NamespaceSummary, error) {
	return nil, domain.ErrNotImplemented
}

func (s *Service) TopWorkflowTypes(context.Context, string, int) ([]domain.WorkflowTypeSummary, error) {
	return nil, domain.ErrNotImplemented
}

func (s *Service) WorkflowUsage(context.Context, string, string) (domain.WorkflowUsage, error) {
	return domain.WorkflowUsage{}, domain.ErrNotImplemented
}

var _ domain.Analyzer = (*Service)(nil)
