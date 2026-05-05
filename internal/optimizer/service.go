package optimizer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"temporal-cost-optimizer/internal/domain"
	"temporal-cost-optimizer/internal/temporalcloud"

	common "go.temporal.io/api/common/v1"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
	"google.golang.org/protobuf/proto"
)

const (
	largePayloadThresholdBytes = 64 * 1024
	excessiveSignalThreshold   = 10
	redundantActivityThreshold = 2
	historyBloatThreshold      = 1000
)

type TemporalHistoryClient interface {
	LastCompletedWorkflowHistory(ctx context.Context, namespace string, workflowID string) (temporalcloud.WorkflowHistory, error)
}

type Service struct {
	temporal TemporalHistoryClient
}

func NewService(temporal TemporalHistoryClient) *Service {
	return &Service{temporal: temporal}
}

func (s *Service) AnalyzeWorkflow(ctx context.Context, namespace string, workflowID string) (domain.WorkflowAnalysis, error) {
	if s.temporal == nil {
		return domain.WorkflowAnalysis{}, domain.ErrNotImplemented
	}

	workflowHistory, err := s.temporal.LastCompletedWorkflowHistory(ctx, namespace, workflowID)
	if err != nil {
		return domain.WorkflowAnalysis{}, err
	}

	findings := analyzeEvents(workflowHistory.Events)
	return domain.WorkflowAnalysis{
		WorkflowID:      workflowID,
		WorkflowRunID:   workflowHistory.RunID,
		Signals:         findings,
		Recommendations: recommendationsFor(findings),
	}, nil
}

var _ domain.Optimizer = (*Service)(nil)

func analyzeEvents(events []*history.HistoryEvent) []domain.AnalysisFinding {
	var findings []domain.AnalysisFinding
	largePayloadEvents := 0
	signalCount := 0
	activityCounts := make(map[string]int)
	redundantActivityCount := 0

	for _, event := range events {
		if event == nil {
			continue
		}

		for _, payloads := range payloadsForEvent(event) {
			if payloadsSize(payloads) > largePayloadThresholdBytes {
				largePayloadEvents++
			}
		}

		if event.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED {
			signalCount++
		}

		if event.GetEventType() == enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED {
			attributes := event.GetActivityTaskScheduledEventAttributes()
			if attributes == nil {
				continue
			}
			key := attributes.GetActivityType().GetName() + ":" + payloadFingerprint(attributes.GetInput())
			activityCounts[key]++
			if activityCounts[key] == redundantActivityThreshold {
				redundantActivityCount++
			}
		}
	}

	if largePayloadEvents > 0 {
		findings = append(findings, domain.AnalysisFinding{
			Type:     "large_payload",
			Severity: "high",
			Evidence: fmt.Sprintf("%d events exceed payload threshold", largePayloadEvents),
		})
	}
	if signalCount > excessiveSignalThreshold {
		findings = append(findings, domain.AnalysisFinding{
			Type:     "excessive_signals",
			Severity: "medium",
			Evidence: fmt.Sprintf("%d signals in one execution", signalCount),
		})
	}
	if redundantActivityCount > 0 {
		findings = append(findings, domain.AnalysisFinding{
			Type:     "redundant_activities",
			Severity: "medium",
			Evidence: fmt.Sprintf("%d repeated activity input patterns", redundantActivityCount),
		})
	}
	if len(events) > historyBloatThreshold {
		findings = append(findings, domain.AnalysisFinding{
			Type:     "history_bloat",
			Severity: "medium",
			Evidence: fmt.Sprintf("%d history events in one execution", len(events)),
		})
	}

	return findings
}

func payloadsForEvent(event *history.HistoryEvent) []*common.Payloads {
	switch event.GetEventType() {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		if attributes := event.GetWorkflowExecutionStartedEventAttributes(); attributes != nil {
			return []*common.Payloads{attributes.GetInput()}
		}
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED:
		if attributes := event.GetWorkflowExecutionSignaledEventAttributes(); attributes != nil {
			return []*common.Payloads{attributes.GetInput()}
		}
	case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		if attributes := event.GetActivityTaskScheduledEventAttributes(); attributes != nil {
			return []*common.Payloads{attributes.GetInput()}
		}
	}
	return nil
}

func payloadsSize(payloads *common.Payloads) int {
	total := 0
	for _, payload := range payloads.GetPayloads() {
		total += len(payload.GetData())
		for key, value := range payload.GetMetadata() {
			total += len(key) + len(value)
		}
	}
	return total
}

func payloadFingerprint(payloads *common.Payloads) string {
	if payloads == nil {
		return "nil"
	}
	data, err := proto.Marshal(payloads)
	if err != nil {
		return fmt.Sprintf("unmarshalable:%p", payloads)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func recommendationsFor(findings []domain.AnalysisFinding) []string {
	recommendationByType := map[string]string{
		"large_payload":        "Compress large payloads before storing them in workflow state.",
		"excessive_signals":    "Batch signals where possible.",
		"redundant_activities": "Deduplicate repeated activities using memoization or cached results.",
		"history_bloat":        "Reduce workflow history size with continue-as-new or by moving large state outside workflow history.",
	}

	seen := make(map[string]bool)
	var recommendations []string
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
