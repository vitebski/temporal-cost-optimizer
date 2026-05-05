package optimizer

import (
	"context"
	"errors"
	"testing"

	common "go.temporal.io/api/common/v1"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"

	"temporal-cost-optimizer/internal/domain"
	"temporal-cost-optimizer/internal/temporalcloud"
)

func TestAnalyzeWorkflowDetectsLargePayloads(t *testing.T) {
	client := &fakeHistoryClient{
		history: temporalcloud.WorkflowHistory{
			Namespace:  "payments-prod",
			WorkflowID: "wf-123",
			RunID:      "run-001",
			Events: []*history.HistoryEvent{
				workflowStartedEvent(payloadWithSize(65 * 1024)),
			},
		},
	}
	service := NewService(client)

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}

	assertFinding(t, analysis.Signals, 0, "large_payload", "high")
	assertRecommendation(t, analysis.Recommendations, "Compress large payloads before storing them in workflow state.")
}

func TestAnalyzeWorkflowDetectsExcessiveSignals(t *testing.T) {
	client := &fakeHistoryClient{
		history: temporalcloud.WorkflowHistory{
			WorkflowID: "wf-123",
			RunID:      "run-001",
			Events:     repeatSignals(11),
		},
	}
	service := NewService(client)

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}

	assertFinding(t, analysis.Signals, 0, "excessive_signals", "medium")
	assertRecommendation(t, analysis.Recommendations, "Batch signals where possible.")
}

func TestAnalyzeWorkflowDetectsRedundantActivities(t *testing.T) {
	events := []*history.HistoryEvent{
		activityScheduledEvent("ChargeCard", payloadWithSize(128)),
		activityScheduledEvent("ChargeCard", payloadWithSize(128)),
	}
	client := &fakeHistoryClient{
		history: temporalcloud.WorkflowHistory{
			WorkflowID: "wf-123",
			RunID:      "run-001",
			Events:     events,
		},
	}
	service := NewService(client)

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}

	assertFinding(t, analysis.Signals, 0, "redundant_activities", "medium")
	assertRecommendation(t, analysis.Recommendations, "Deduplicate repeated activities using memoization or cached results.")
}

func TestAnalyzeWorkflowDetectsHistoryBloat(t *testing.T) {
	client := &fakeHistoryClient{
		history: temporalcloud.WorkflowHistory{
			WorkflowID: "wf-123",
			RunID:      "run-001",
			Events:     makeGenericEvents(1001),
		},
	}
	service := NewService(client)

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}

	assertFinding(t, analysis.Signals, 0, "history_bloat", "medium")
	assertRecommendation(t, analysis.Recommendations, "Reduce workflow history size with continue-as-new or by moving large state outside workflow history.")
}

func TestAnalyzeWorkflowReturnsNoFindingsForSmallHistory(t *testing.T) {
	client := &fakeHistoryClient{
		history: temporalcloud.WorkflowHistory{
			WorkflowID: "wf-123",
			RunID:      "run-001",
			Events: []*history.HistoryEvent{
				workflowStartedEvent(payloadWithSize(128)),
				signalEvent("Update", payloadWithSize(64)),
				activityScheduledEvent("ChargeCard", payloadWithSize(128)),
			},
		},
	}
	service := NewService(client)

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}

	if client.namespace != "payments-prod" {
		t.Fatalf("namespace = %q, want payments-prod", client.namespace)
	}
	if client.workflowID != "wf-123" {
		t.Fatalf("workflow ID = %q, want wf-123", client.workflowID)
	}
	if analysis.WorkflowID != "wf-123" {
		t.Fatalf("analysis workflow ID = %q, want wf-123", analysis.WorkflowID)
	}
	if analysis.WorkflowRunID != "run-001" {
		t.Fatalf("analysis run ID = %q, want run-001", analysis.WorkflowRunID)
	}
	if len(analysis.Signals) != 0 {
		t.Fatalf("findings length = %d, want 0", len(analysis.Signals))
	}
	if len(analysis.Recommendations) != 0 {
		t.Fatalf("recommendations length = %d, want 0", len(analysis.Recommendations))
	}
}

func TestAnalyzeWorkflowPropagatesHistoryErrors(t *testing.T) {
	wantErr := errors.New("history unavailable")
	service := NewService(&fakeHistoryClient{err: wantErr})

	_, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

type fakeHistoryClient struct {
	namespace  string
	workflowID string
	history    temporalcloud.WorkflowHistory
	err        error
}

func (f *fakeHistoryClient) LastCompletedWorkflowHistory(_ context.Context, namespace string, workflowID string) (temporalcloud.WorkflowHistory, error) {
	f.namespace = namespace
	f.workflowID = workflowID
	if f.err != nil {
		return temporalcloud.WorkflowHistory{}, f.err
	}
	return f.history, nil
}

func assertFinding(t *testing.T, findings []domain.AnalysisFinding, index int, findingType string, severity string) {
	t.Helper()

	if len(findings) <= index {
		t.Fatalf("findings length = %d, want at least %d", len(findings), index+1)
	}
	if findings[index].Type != findingType {
		t.Fatalf("finding type = %q, want %q", findings[index].Type, findingType)
	}
	if findings[index].Severity != severity {
		t.Fatalf("finding severity = %q, want %q", findings[index].Severity, severity)
	}
	if findings[index].Evidence == "" {
		t.Fatal("finding evidence is empty")
	}
}

func assertRecommendation(t *testing.T, recommendations []string, recommendation string) {
	t.Helper()

	for _, got := range recommendations {
		if got == recommendation {
			return
		}
	}
	t.Fatalf("recommendations = %v, want %q", recommendations, recommendation)
}

func workflowStartedEvent(input *common.Payloads) *history.HistoryEvent {
	return &history.HistoryEvent{
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
				Input: input,
			},
		},
	}
}

func signalEvent(signalName string, input *common.Payloads) *history.HistoryEvent {
	return &history.HistoryEvent{
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED,
		Attributes: &history.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
			WorkflowExecutionSignaledEventAttributes: &history.WorkflowExecutionSignaledEventAttributes{
				SignalName: signalName,
				Input:      input,
			},
		},
	}
}

func activityScheduledEvent(activityName string, input *common.Payloads) *history.HistoryEvent {
	return &history.HistoryEvent{
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
				ActivityType: &common.ActivityType{Name: activityName},
				Input:        input,
			},
		},
	}
}

func payloadWithSize(size int) *common.Payloads {
	return &common.Payloads{
		Payloads: []*common.Payload{
			{Data: make([]byte, size)},
		},
	}
}

func repeatSignals(count int) []*history.HistoryEvent {
	events := make([]*history.HistoryEvent, 0, count)
	for i := 0; i < count; i++ {
		events = append(events, signalEvent("Update", payloadWithSize(64)))
	}
	return events
}

func makeGenericEvents(count int) []*history.HistoryEvent {
	events := make([]*history.HistoryEvent, count)
	for i := range events {
		events[i] = &history.HistoryEvent{EventId: int64(i + 1)}
	}
	return events
}
