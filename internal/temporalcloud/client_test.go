package temporalcloud

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	common "go.temporal.io/api/common/v1"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
	workflow "go.temporal.io/api/workflow/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	cloudnamespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	usage "go.temporal.io/cloud-sdk/api/usage/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/domain"
)

func TestGetUsageBuildsCloudSDKRequest(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	fake := &fakeCloudService{
		usageResponse: &cloudservice.GetUsageResponse{
			Summaries: []*usage.Summary{{Incomplete: true}},
		},
	}
	client := newClientForUsageService(config.TemporalConfig{UsagePageSize: 100}, fake, nil)

	page, err := client.GetUsage(context.Background(), UsageQuery{
		StartTimeInclusive: start,
		EndTimeExclusive:   end,
		PageSize:           25,
		PageToken:          "next-token",
	})
	if err != nil {
		t.Fatalf("GetUsage returned error: %v", err)
	}

	if fake.usageRequest.GetStartTimeInclusive().AsTime() != start {
		t.Fatalf("start time = %s, want %s", fake.usageRequest.GetStartTimeInclusive().AsTime(), start)
	}
	if fake.usageRequest.GetEndTimeExclusive().AsTime() != end {
		t.Fatalf("end time = %s, want %s", fake.usageRequest.GetEndTimeExclusive().AsTime(), end)
	}
	if fake.usageRequest.GetPageSize() != 25 {
		t.Fatalf("page size = %d, want 25", fake.usageRequest.GetPageSize())
	}
	if fake.usageRequest.GetPageToken() != "next-token" {
		t.Fatalf("page token = %q, want next-token", fake.usageRequest.GetPageToken())
	}
	if len(page.Summaries) != 1 {
		t.Fatalf("summaries length = %d, want 1", len(page.Summaries))
	}
}

func TestGetUsageUsesConfiguredPageSize(t *testing.T) {
	fake := &fakeCloudService{usageResponse: &cloudservice.GetUsageResponse{}}
	client := newClientForUsageService(config.TemporalConfig{UsagePageSize: 300}, fake, nil)

	_, err := client.GetUsage(context.Background(), UsageQuery{})
	if err != nil {
		t.Fatalf("GetUsage returned error: %v", err)
	}

	if fake.usageRequest.GetPageSize() != 300 {
		t.Fatalf("page size = %d, want configured default 300", fake.usageRequest.GetPageSize())
	}
}

func TestLastCompletedWorkflowHistoryListsCompletedExecutionsByNamespace(t *testing.T) {
	closeTime := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	workflowFake := &fakeWorkflowService{
		listResponses: []*workflowservice.ListWorkflowExecutionsResponse{
			{
				Executions: []*workflow.WorkflowExecutionInfo{
					workflowExecution("wf-123", "run-001", closeTime),
				},
			},
		},
		historyResponses: []*workflowservice.GetWorkflowExecutionHistoryResponse{
			historyResponse(historyEvent(1), nil),
		},
	}
	cloudFake := &fakeCloudService{namespaceResponse: namespaceResponse("payments-prod.namespace.tmprl.cloud:7233")}
	workflowFactory := &fakeWorkflowServiceFactory{service: workflowFake}
	client := newClientForServices(config.TemporalConfig{}, cloudFake, workflowFactory, nil)

	workflowHistory, err := client.LastCompletedWorkflowHistory(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("LastCompletedWorkflowHistory returned error: %v", err)
	}

	if cloudFake.namespaceRequest.GetNamespace() != "payments-prod" {
		t.Fatalf("namespace lookup = %q, want payments-prod", cloudFake.namespaceRequest.GetNamespace())
	}
	if workflowFactory.endpoint != "payments-prod.namespace.tmprl.cloud:7233" {
		t.Fatalf("workflow endpoint = %q, want payments-prod.namespace.tmprl.cloud:7233", workflowFactory.endpoint)
	}
	if workflowHistory.Namespace != "payments-prod" {
		t.Fatalf("namespace = %q, want payments-prod", workflowHistory.Namespace)
	}
	if workflowHistory.WorkflowID != "wf-123" {
		t.Fatalf("workflow ID = %q, want wf-123", workflowHistory.WorkflowID)
	}
	if workflowHistory.RunID != "run-001" {
		t.Fatalf("run ID = %q, want run-001", workflowHistory.RunID)
	}
	if len(workflowHistory.Events) != 1 {
		t.Fatalf("events length = %d, want 1", len(workflowHistory.Events))
	}
	if workflowFake.listRequests[0].GetNamespace() != "payments-prod" {
		t.Fatalf("list namespace = %q, want payments-prod", workflowFake.listRequests[0].GetNamespace())
	}
	if workflowFake.listRequests[0].GetPageSize() != defaultWorkflowListPageSize {
		t.Fatalf("list page size = %d, want %d", workflowFake.listRequests[0].GetPageSize(), defaultWorkflowListPageSize)
	}
	if !strings.Contains(workflowFake.listRequests[0].GetQuery(), `WorkflowId = "wf-123"`) {
		t.Fatalf("list query = %q, want workflow ID filter", workflowFake.listRequests[0].GetQuery())
	}
	if !strings.Contains(workflowFake.listRequests[0].GetQuery(), `ExecutionStatus = "Completed"`) {
		t.Fatalf("list query = %q, want completed status filter", workflowFake.listRequests[0].GetQuery())
	}
	if workflowFake.historyRequests[0].GetNamespace() != "payments-prod" {
		t.Fatalf("history namespace = %q, want payments-prod", workflowFake.historyRequests[0].GetNamespace())
	}
	if workflowFake.historyRequests[0].GetExecution().GetWorkflowId() != "wf-123" {
		t.Fatalf("history workflow ID = %q, want wf-123", workflowFake.historyRequests[0].GetExecution().GetWorkflowId())
	}
	if workflowFake.historyRequests[0].GetExecution().GetRunId() != "run-001" {
		t.Fatalf("history run ID = %q, want run-001", workflowFake.historyRequests[0].GetExecution().GetRunId())
	}
	if workflowFake.historyRequests[0].GetMaximumPageSize() != defaultHistoryPageSize {
		t.Fatalf("history page size = %d, want %d", workflowFake.historyRequests[0].GetMaximumPageSize(), defaultHistoryPageSize)
	}
}

func TestLastCompletedWorkflowHistorySelectsMostRecentCompletedRun(t *testing.T) {
	workflowFake := &fakeWorkflowService{
		listResponses: []*workflowservice.ListWorkflowExecutionsResponse{
			{
				Executions: []*workflow.WorkflowExecutionInfo{
					workflowExecution("wf-123", "run-old", time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)),
					workflowExecution("wf-123", "run-new", time.Date(2026, 5, 5, 11, 0, 0, 0, time.UTC)),
				},
			},
		},
		historyResponses: []*workflowservice.GetWorkflowExecutionHistoryResponse{
			historyResponse(historyEvent(1), nil),
		},
	}
	client := newClientForServices(config.TemporalConfig{}, &fakeCloudService{namespaceResponse: namespaceResponse("payments-prod.namespace.tmprl.cloud:7233")}, &fakeWorkflowServiceFactory{service: workflowFake}, nil)

	workflowHistory, err := client.LastCompletedWorkflowHistory(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("LastCompletedWorkflowHistory returned error: %v", err)
	}

	if workflowHistory.RunID != "run-new" {
		t.Fatalf("run ID = %q, want run-new", workflowHistory.RunID)
	}
}

func TestLastCompletedWorkflowHistoryFetchesAllHistoryPages(t *testing.T) {
	workflowFake := &fakeWorkflowService{
		listResponses: []*workflowservice.ListWorkflowExecutionsResponse{
			{
				Executions: []*workflow.WorkflowExecutionInfo{
					workflowExecution("wf-123", "run-001", time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		historyResponses: []*workflowservice.GetWorkflowExecutionHistoryResponse{
			historyResponse(historyEvent(1), []byte("page-2")),
			historyResponse(historyEvent(2), nil),
		},
	}
	client := newClientForServices(config.TemporalConfig{}, &fakeCloudService{namespaceResponse: namespaceResponse("payments-prod.namespace.tmprl.cloud:7233")}, &fakeWorkflowServiceFactory{service: workflowFake}, nil)

	workflowHistory, err := client.LastCompletedWorkflowHistory(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("LastCompletedWorkflowHistory returned error: %v", err)
	}

	if len(workflowFake.historyRequests) != 2 {
		t.Fatalf("history requests = %d, want 2", len(workflowFake.historyRequests))
	}
	if string(workflowFake.historyRequests[1].GetNextPageToken()) != "page-2" {
		t.Fatalf("second page token = %q, want page-2", string(workflowFake.historyRequests[1].GetNextPageToken()))
	}
	if len(workflowHistory.Events) != 2 {
		t.Fatalf("events length = %d, want 2", len(workflowHistory.Events))
	}
}

func TestLastCompletedWorkflowHistoryReturnsNotImplementedWithoutNamespaceEndpoint(t *testing.T) {
	client := newClientForUsageService(config.TemporalConfig{}, nil, nil)

	_, err := client.LastCompletedWorkflowHistory(context.Background(), "payments-prod", "wf-123")
	if !errors.Is(err, domain.ErrNotImplemented) {
		t.Fatalf("error = %v, want ErrNotImplemented", err)
	}
}

func TestLastCompletedWorkflowHistoryReturnsErrorWhenNoCompletedRunExists(t *testing.T) {
	workflowFake := &fakeWorkflowService{
		listResponses: []*workflowservice.ListWorkflowExecutionsResponse{
			{},
		},
	}
	client := newClientForServices(config.TemporalConfig{}, &fakeCloudService{namespaceResponse: namespaceResponse("payments-prod.namespace.tmprl.cloud:7233")}, &fakeWorkflowServiceFactory{service: workflowFake}, nil)

	_, err := client.LastCompletedWorkflowHistory(context.Background(), "payments-prod", "wf-123")
	if err == nil {
		t.Fatal("error = nil, want no completed workflow error")
	}
	if !strings.Contains(err.Error(), "no completed workflow execution found") {
		t.Fatalf("error = %q, want no completed workflow detail", err.Error())
	}
}

type fakeCloudService struct {
	usageRequest      *cloudservice.GetUsageRequest
	usageResponse     *cloudservice.GetUsageResponse
	namespaceRequest  *cloudservice.GetNamespaceRequest
	namespaceResponse *cloudservice.GetNamespaceResponse
	namespaceErr      error
}

func (f *fakeCloudService) GetUsage(ctx context.Context, req *cloudservice.GetUsageRequest, _ ...grpc.CallOption) (*cloudservice.GetUsageResponse, error) {
	f.usageRequest = req
	return f.usageResponse, nil
}

func (f *fakeCloudService) GetNamespace(_ context.Context, req *cloudservice.GetNamespaceRequest, _ ...grpc.CallOption) (*cloudservice.GetNamespaceResponse, error) {
	f.namespaceRequest = req
	if f.namespaceErr != nil {
		return nil, f.namespaceErr
	}
	return f.namespaceResponse, nil
}

type fakeWorkflowServiceFactory struct {
	endpoint string
	service  workflowService
	closer   closer
	err      error
}

func (f *fakeWorkflowServiceFactory) NewWorkflowService(_ config.TemporalConfig, endpoint string) (workflowService, closer, error) {
	f.endpoint = endpoint
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.service, f.closer, nil
}

type fakeWorkflowService struct {
	listRequests     []*workflowservice.ListWorkflowExecutionsRequest
	listResponses    []*workflowservice.ListWorkflowExecutionsResponse
	listErr          error
	historyRequests  []*workflowservice.GetWorkflowExecutionHistoryRequest
	historyResponses []*workflowservice.GetWorkflowExecutionHistoryResponse
	historyErr       error
}

func (f *fakeWorkflowService) ListWorkflowExecutions(_ context.Context, req *workflowservice.ListWorkflowExecutionsRequest, _ ...grpc.CallOption) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	f.listRequests = append(f.listRequests, req)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if len(f.listResponses) == 0 {
		return &workflowservice.ListWorkflowExecutionsResponse{}, nil
	}
	resp := f.listResponses[0]
	f.listResponses = f.listResponses[1:]
	return resp, nil
}

func (f *fakeWorkflowService) GetWorkflowExecutionHistory(_ context.Context, req *workflowservice.GetWorkflowExecutionHistoryRequest, _ ...grpc.CallOption) (*workflowservice.GetWorkflowExecutionHistoryResponse, error) {
	f.historyRequests = append(f.historyRequests, req)
	if f.historyErr != nil {
		return nil, f.historyErr
	}
	if len(f.historyResponses) == 0 {
		return &workflowservice.GetWorkflowExecutionHistoryResponse{}, nil
	}
	resp := f.historyResponses[0]
	f.historyResponses = f.historyResponses[1:]
	return resp, nil
}

func workflowExecution(workflowID string, runID string, closeTime time.Time) *workflow.WorkflowExecutionInfo {
	return &workflow.WorkflowExecutionInfo{
		Execution: &common.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		Status:    enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
		CloseTime: timestamppb.New(closeTime),
	}
}

func historyResponse(event *history.HistoryEvent, nextPageToken []byte) *workflowservice.GetWorkflowExecutionHistoryResponse {
	return &workflowservice.GetWorkflowExecutionHistoryResponse{
		History: &history.History{
			Events: []*history.HistoryEvent{event},
		},
		NextPageToken: nextPageToken,
	}
}

func historyEvent(eventID int64) *history.HistoryEvent {
	return &history.HistoryEvent{EventId: eventID}
}

func namespaceResponse(grpcAddress string) *cloudservice.GetNamespaceResponse {
	return &cloudservice.GetNamespaceResponse{
		Namespace: &cloudnamespace.Namespace{
			Endpoints: &cloudnamespace.Endpoints{
				GrpcAddress: grpcAddress,
			},
		},
	}
}
