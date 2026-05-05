package sampledata

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"temporal-cost-optimizer/internal/domain"
)

type Service struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func NewService() *Service {
	return &Service{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (s *Service) TopNamespaces(_ context.Context, top int) ([]domain.NamespaceSummary, error) {
	count := clampTop(top, len(sampleNamespaces))
	items := make([]domain.NamespaceSummary, 0, count)
	for i := 0; i < count; i++ {
		usageScore := s.float(18000, 140000)
		activeStorage := s.float(100, 9000)
		retainedStorage := s.float(500, 30000)
		items = append(items, domain.NamespaceSummary{
			Namespace:     sampleNamespaces[i],
			Rank:          i + 1,
			UsageScore:    roundTo2DP(usageScore),
			EstimatedCost: roundTo2DP(usageScore * 0.0001),
			Storage: domain.StorageBreakdown{
				Active: domain.StorageUsage{
					Usage: roundTo2DP(activeStorage),
					Cost:  roundTo2DP(activeStorage * 0.00002),
				},
				Retained: domain.StorageUsage{
					Usage: roundTo2DP(retainedStorage),
					Cost:  roundTo2DP(retainedStorage * 0.00001),
				},
			},
			Trend:      s.choice([]string{"up", "down", "flat"}),
			Incomplete: false,
		})
	}
	return items, nil
}

func (s *Service) TopWorkflowTypes(_ context.Context, namespace string, top int) ([]domain.WorkflowTypeSummary, error) {
	count := clampTop(top, len(sampleWorkflowTypes))
	items := make([]domain.WorkflowTypeSummary, 0, count)
	for i := 0; i < count; i++ {
		usageScore := s.float(5000, 60000)
		activeStorage := s.float(10, 1500)
		retainedStorage := s.float(50, 6000)
		items = append(items, domain.WorkflowTypeSummary{
			Namespace:     namespace,
			WorkflowType:  sampleWorkflowTypes[i],
			UsageScore:    roundTo2DP(usageScore),
			EstimatedCost: roundTo2DP(usageScore * 0.0001),
			Storage: domain.StorageBreakdown{
				Active: domain.StorageUsage{
					Usage: roundTo2DP(activeStorage),
					Cost:  roundTo2DP(activeStorage * 0.00002),
				},
				Retained: domain.StorageUsage{
					Usage: roundTo2DP(retainedStorage),
					Cost:  roundTo2DP(retainedStorage * 0.00001),
				},
			},
			Executions: s.int(20, 2500),
			Signals:    s.int(5, 900),
			Activities: s.int(50, 9000),
		})
	}
	return items, nil
}

func (s *Service) WorkflowUsage(_ context.Context, namespace string, workflowType string) (domain.WorkflowUsage, error) {
	activeStorage := s.float(20, 1800)
	retainedStorage := s.float(100, 9000)
	executions := s.int(25, 2500)
	avgHistory := s.int(60, 700)

	return domain.WorkflowUsage{
		Namespace:    namespace,
		WorkflowType: workflowType,
		Summary: domain.WorkflowUsageSummary{
			Storage: domain.StorageBreakdown{
				Active: domain.StorageUsage{
					Usage: roundTo2DP(activeStorage),
					Cost:  roundTo2DP(activeStorage * 0.00002),
				},
				Retained: domain.StorageUsage{
					Usage: roundTo2DP(retainedStorage),
					Cost:  roundTo2DP(retainedStorage * 0.00001),
				},
			},
			Executions:       executions,
			BillableActions:  executions * s.int(20, 250),
			AvgHistoryEvents: avgHistory,
			P95HistoryEvents: avgHistory + s.int(100, 900),
		},
	}, nil
}

func (s *Service) AnalyzeWorkflow(_ context.Context, _ string, workflowID string) (domain.WorkflowAnalysis, error) {
	findings := s.sampleFindings()
	return domain.WorkflowAnalysis{
		WorkflowID:      workflowID,
		WorkflowRunID:   fmt.Sprintf("sample-run-%06d", s.int(1, 999999)),
		Signals:         findings,
		Recommendations: recommendationsFor(findings),
	}, nil
}

var _ domain.Analyzer = (*Service)(nil)
var _ domain.Optimizer = (*Service)(nil)

var sampleNamespaces = []string{
	"rcp-k8s-staging-cluster-upgrade.nemly",
	"payments-prod",
	"orders-prod",
	"identity-prod",
	"billing-prod",
	"search-prod",
	"fraud-prod",
}

var sampleWorkflowTypes = []string{
	"SingleClusterUpgradeWorkflow",
	"ChargeCardWorkflow",
	"FulfillOrderWorkflow",
	"InvoiceCustomerWorkflow",
	"ProvisionAccountWorkflow",
	"ReconcileLedgerWorkflow",
}

var sampleFindingTemplates = []domain.AnalysisFinding{
	{
		Type:     "large_payload",
		Severity: "high",
		Evidence: "sample: payloads between 80KB and 2MB were observed in workflow history",
	},
	{
		Type:     "excessive_signals",
		Severity: "medium",
		Evidence: "sample: one execution received more than 25 signals",
	},
	{
		Type:     "redundant_activities",
		Severity: "medium",
		Evidence: "sample: repeated activity inputs suggest cacheable work",
	},
	{
		Type:     "history_bloat",
		Severity: "medium",
		Evidence: "sample: execution history exceeded 1,000 events",
	},
}

func (s *Service) sampleFindings() []domain.AnalysisFinding {
	count := s.int(1, len(sampleFindingTemplates))
	offset := s.int(0, len(sampleFindingTemplates)-1)
	findings := make([]domain.AnalysisFinding, 0, count)
	for i := 0; i < count; i++ {
		findings = append(findings, sampleFindingTemplates[(offset+i)%len(sampleFindingTemplates)])
	}
	return findings
}

func recommendationsFor(findings []domain.AnalysisFinding) []string {
	recommendationByType := map[string]string{
		"large_payload":        "Compress large payloads before storing them in workflow state.",
		"excessive_signals":    "Batch signals where possible.",
		"redundant_activities": "Deduplicate repeated activities using memoization or cached results.",
		"history_bloat":        "Reduce workflow history size with continue-as-new or by moving large state outside workflow history.",
	}
	recommendations := make([]string, 0, len(findings))
	seen := make(map[string]bool)
	for _, finding := range findings {
		recommendation := recommendationByType[finding.Type]
		if recommendation == "" || seen[recommendation] {
			continue
		}
		seen[recommendation] = true
		recommendations = append(recommendations, recommendation)
	}
	return recommendations
}

func clampTop(top int, max int) int {
	if top < 1 {
		return 5
	}
	if top > max {
		return max
	}
	return top
}

func roundTo2DP(value float64) float64 {
	return math.Round(value*100) / 100
}

func (s *Service) int(min int, max int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return min + s.rng.Intn(max-min+1)
}

func (s *Service) float(min float64, max float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return min + s.rng.Float64()*(max-min)
}

func (s *Service) choice(values []string) string {
	return values[s.int(0, len(values)-1)]
}
