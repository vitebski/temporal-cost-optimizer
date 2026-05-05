package analyzer

import (
	"context"
	"math"
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

// estimatedCostPerScoreUnit is a placeholder multiplier so the namespaces
// screen has a non-zero cost column for the demo. It is not a real billing
// rate and intentionally lives in code rather than configuration.
const estimatedCostPerScoreUnit = 0.0001

// byteSecondsToGBHours converts the Cloud Usage API's storage unit
// (bytes × seconds) into the much more readable GB-hours.
func byteSecondsToGBHours(byteSeconds float64) float64 {
	return byteSeconds / (1e9 * 3600)
}

// roundTo2DP rounds a value to two decimal places to keep the JSON
// response readable for the extension UI.
func roundTo2DP(value float64) float64 {
	return math.Round(value*100) / 100
}

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
					gbHours := byteSecondsToGBHours(record.GetValue())
					item.Storage.Active.Usage += gbHours
					item.UsageScore += gbHours
				case usage.RecordType_RECORD_TYPE_RETAINED_STORAGE:
					gbHours := byteSecondsToGBHours(record.GetValue())
					item.Storage.Retained.Usage += gbHours
					item.UsageScore += gbHours
				}
			}
		}
	}

	items := make([]domain.NamespaceSummary, 0, len(byNamespace))
	for _, item := range byNamespace {
		item.EstimatedCost = item.UsageScore * estimatedCostPerScoreUnit

		item.UsageScore = roundTo2DP(item.UsageScore)
		item.EstimatedCost = roundTo2DP(item.EstimatedCost)
		item.Storage.Active.Usage = roundTo2DP(item.Storage.Active.Usage)
		item.Storage.Active.Cost = roundTo2DP(item.Storage.Active.Cost)
		item.Storage.Retained.Usage = roundTo2DP(item.Storage.Retained.Usage)
		item.Storage.Retained.Cost = roundTo2DP(item.Storage.Retained.Cost)

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
