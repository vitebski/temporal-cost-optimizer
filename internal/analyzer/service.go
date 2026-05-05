package analyzer

import (
	"context"
	"sort"

	"temporal-cost-optimizer/internal/domain"
	"temporal-cost-optimizer/internal/temporalcloud"

	usage "go.temporal.io/cloud-sdk/api/usage/v1"
)

type TemporalUsageClient interface {
	GetUsage(context.Context, temporalcloud.UsageQuery) (temporalcloud.UsagePage, error)
}

type Service struct {
	temporal TemporalUsageClient
}

func NewService(temporal TemporalUsageClient) *Service {
	return &Service{temporal: temporal}
}

func (s *Service) TopNamespaces(ctx context.Context, top int) ([]domain.NamespaceSummary, error) {
	var summaries []*usage.Summary
	pageToken := ""
	for {
		page, err := s.temporal.GetUsage(ctx, temporalcloud.UsageQuery{PageToken: pageToken})
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, page.Summaries...)
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}

	items := namespaceSummaries(summaries)
	sort.Slice(items, func(i, j int) bool {
		if items[i].UsageScore == items[j].UsageScore {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].UsageScore > items[j].UsageScore
	})

	if top > 0 && len(items) > top {
		items = items[:top]
	}
	for i := range items {
		items[i].Rank = i + 1
	}

	return items, nil
}

func (s *Service) TopWorkflowTypes(context.Context, string, int) ([]domain.WorkflowTypeSummary, error) {
	return nil, domain.ErrNotImplemented
}

func (s *Service) WorkflowUsage(context.Context, string, string) (domain.WorkflowUsage, error) {
	return domain.WorkflowUsage{}, domain.ErrNotImplemented
}

var _ domain.Analyzer = (*Service)(nil)

func namespaceSummaries(summaries []*usage.Summary) []domain.NamespaceSummary {
	byNamespace := make(map[string]*domain.NamespaceSummary)
	for _, summary := range summaries {
		for _, group := range summary.GetRecordGroups() {
			namespace := namespaceForGroup(group)
			if namespace == "" {
				continue
			}

			item, ok := byNamespace[namespace]
			if !ok {
				item = &domain.NamespaceSummary{Namespace: namespace}
				byNamespace[namespace] = item
			}
			item.Incomplete = item.Incomplete || summary.GetIncomplete()

			for _, record := range group.GetRecords() {
				switch record.GetType() {
				case usage.RecordType_RECORD_TYPE_ACTIONS:
					item.UsageScore += record.GetValue()
				case usage.RecordType_RECORD_TYPE_ACTIVE_STORAGE:
					item.Storage.Active.Usage += record.GetValue()
					item.UsageScore += record.GetValue()
				case usage.RecordType_RECORD_TYPE_RETAINED_STORAGE:
					item.Storage.Retained.Usage += record.GetValue()
					item.UsageScore += record.GetValue()
				}
			}
		}
	}

	items := make([]domain.NamespaceSummary, 0, len(byNamespace))
	for _, item := range byNamespace {
		items = append(items, *item)
	}
	return items
}

func namespaceForGroup(group *usage.RecordGroup) string {
	for _, groupBy := range group.GetGroupBys() {
		if groupBy.GetKey() == usage.GroupByKey_GROUP_BY_KEY_NAMESPACE {
			return groupBy.GetValue()
		}
	}
	return ""
}
