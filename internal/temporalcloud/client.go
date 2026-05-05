package temporalcloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	common "go.temporal.io/api/common/v1"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	usage "go.temporal.io/cloud-sdk/api/usage/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"

	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/domain"
)

const (
	defaultHistoryPageSize      int32 = 1000
	defaultWorkflowListPageSize int32 = 100
)

type UsageQuery struct {
	StartTimeInclusive time.Time
	EndTimeExclusive   time.Time
	PageSize           int32
	PageToken          string
}

type UsagePage struct {
	Summaries     []*usage.Summary
	NextPageToken string
}

type WorkflowHistory struct {
	Namespace  string
	WorkflowID string
	RunID      string
	Events     []*history.HistoryEvent
}

type cloudService interface {
	GetUsage(context.Context, *cloudservice.GetUsageRequest, ...grpc.CallOption) (*cloudservice.GetUsageResponse, error)
	GetNamespace(context.Context, *cloudservice.GetNamespaceRequest, ...grpc.CallOption) (*cloudservice.GetNamespaceResponse, error)
}

type workflowService interface {
	ListWorkflowExecutions(context.Context, *workflowservice.ListWorkflowExecutionsRequest, ...grpc.CallOption) (*workflowservice.ListWorkflowExecutionsResponse, error)
	GetWorkflowExecutionHistory(context.Context, *workflowservice.GetWorkflowExecutionHistoryRequest, ...grpc.CallOption) (*workflowservice.GetWorkflowExecutionHistoryResponse, error)
}

type workflowServiceFactory interface {
	NewWorkflowService(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error)
}

type workflowServiceFactoryFunc func(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error)

func (f workflowServiceFactoryFunc) NewWorkflowService(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error) {
	return f(cfg, endpoint)
}

type closer interface {
	Close() error
}

type Client struct {
	config          config.TemporalConfig
	cloudService    cloudService
	workflowFactory workflowServiceFactory
	closer          closer
}

func NewClient(cfg config.TemporalConfig) (*Client, error) {
	sdkClient, err := cloudclient.New(cloudclient.Options{
		APIKey:     cfg.APIKey,
		HostPort:   cfg.APIHostPort,
		APIVersion: cfg.APIVersion,
		UserAgent:  "temporal-cost-optimizer",
	})
	if err != nil {
		return nil, err
	}

	return newClientForServices(cfg, sdkClient.CloudService(), workflowServiceFactoryFunc(newWorkflowService), sdkClient), nil
}

func newClientForUsageService(cfg config.TemporalConfig, cloudService cloudService, closer closer) *Client {
	return newClientForServices(cfg, cloudService, workflowServiceFactoryFunc(newWorkflowService), closer)
}

func newClientForServices(cfg config.TemporalConfig, cloudService cloudService, workflowFactory workflowServiceFactory, closer closer) *Client {
	return &Client{
		config:          cfg,
		cloudService:    cloudService,
		workflowFactory: workflowFactory,
		closer:          closer,
	}
}

func newWorkflowService(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error) {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(apiKeyCredentials{apiKey: cfg.APIKey}),
		grpc.WithUserAgent("temporal-cost-optimizer"),
	)
	if err != nil {
		return nil, nil, err
	}

	return workflowservice.NewWorkflowServiceClient(conn), conn, nil
}

func (c *Client) Config() config.TemporalConfig {
	return c.config
}

func (c *Client) GetUsage(ctx context.Context, query UsageQuery) (UsagePage, error) {
	req := &cloudservice.GetUsageRequest{
		PageSize:  query.PageSize,
		PageToken: query.PageToken,
	}
	if req.PageSize == 0 {
		req.PageSize = c.config.UsagePageSize
	}
	if !query.StartTimeInclusive.IsZero() {
		req.StartTimeInclusive = timestamppb.New(query.StartTimeInclusive)
	}
	if !query.EndTimeExclusive.IsZero() {
		req.EndTimeExclusive = timestamppb.New(query.EndTimeExclusive)
	}

	resp, err := c.cloudService.GetUsage(ctx, req)
	if err != nil {
		return UsagePage{}, err
	}
	if resp == nil {
		return UsagePage{}, nil
	}

	return UsagePage{
		Summaries:     resp.GetSummaries(),
		NextPageToken: resp.GetNextPageToken(),
	}, nil
}

func (c *Client) LastCompletedWorkflowHistory(ctx context.Context, namespace string, workflowID string) (WorkflowHistory, error) {
	if c.cloudService == nil || c.workflowFactory == nil {
		return WorkflowHistory{}, domain.ErrNotImplemented
	}

	endpoint, err := c.workflowEndpoint(ctx, namespace)
	if err != nil {
		return WorkflowHistory{}, err
	}

	workflowService, workflowCloser, err := c.workflowFactory.NewWorkflowService(c.config, endpoint)
	if err != nil {
		return WorkflowHistory{}, err
	}
	if workflowCloser != nil {
		defer workflowCloser.Close()
	}

	runID, err := c.lastCompletedRunID(ctx, workflowService, namespace, workflowID)
	if err != nil {
		return WorkflowHistory{}, err
	}

	events, err := c.workflowHistory(ctx, workflowService, namespace, workflowID, runID)
	if err != nil {
		return WorkflowHistory{}, err
	}

	return WorkflowHistory{
		Namespace:  namespace,
		WorkflowID: workflowID,
		RunID:      runID,
		Events:     events,
	}, nil
}

func (c *Client) workflowEndpoint(ctx context.Context, namespace string) (string, error) {
	resp, err := c.cloudService.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: namespace})
	if err != nil {
		return "", err
	}

	endpoint := resp.GetNamespace().GetEndpoints().GetGrpcAddress()
	if endpoint == "" {
		return "", domain.ErrNotImplemented
	}

	return endpoint, nil
}

func (c *Client) lastCompletedRunID(ctx context.Context, workflowService workflowService, namespace string, workflowID string) (string, error) {
	var selectedRunID string
	var selectedCloseTime time.Time
	nextPageToken := []byte(nil)

	for {
		resp, err := workflowService.ListWorkflowExecutions(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      defaultWorkflowListPageSize,
			NextPageToken: nextPageToken,
			Query:         completedWorkflowQuery(workflowID),
		})
		if err != nil {
			return "", err
		}

		for _, execution := range resp.GetExecutions() {
			if execution.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_COMPLETED {
				continue
			}
			if execution.GetExecution().GetWorkflowId() != workflowID || execution.GetExecution().GetRunId() == "" {
				continue
			}

			closeTime := execution.GetCloseTime().AsTime()
			if selectedRunID == "" || closeTime.After(selectedCloseTime) {
				selectedRunID = execution.GetExecution().GetRunId()
				selectedCloseTime = closeTime
			}
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			break
		}
	}

	if selectedRunID == "" {
		return "", fmt.Errorf("no completed workflow execution found for namespace %q workflow %q", namespace, workflowID)
	}

	return selectedRunID, nil
}

func (c *Client) workflowHistory(ctx context.Context, workflowService workflowService, namespace string, workflowID string, runID string) ([]*history.HistoryEvent, error) {
	var events []*history.HistoryEvent
	nextPageToken := []byte(nil)

	for {
		resp, err := workflowService.GetWorkflowExecutionHistory(ctx, &workflowservice.GetWorkflowExecutionHistoryRequest{
			Namespace: namespace,
			Execution: &common.WorkflowExecution{
				WorkflowId: workflowID,
				RunId:      runID,
			},
			MaximumPageSize: defaultHistoryPageSize,
			NextPageToken:   nextPageToken,
		})
		if err != nil {
			return nil, err
		}

		events = append(events, resp.GetHistory().GetEvents()...)
		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			break
		}
	}

	return events, nil
}

func completedWorkflowQuery(workflowID string) string {
	escaped := strings.ReplaceAll(strings.ReplaceAll(workflowID, `\`, `\\`), `"`, `\"`)
	return `WorkflowId = "` + escaped + `" AND ExecutionStatus = "Completed"`
}

func (c *Client) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}

type apiKeyCredentials struct {
	apiKey string
}

func (c apiKeyCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + c.apiKey}, nil
}

func (c apiKeyCredentials) RequireTransportSecurity() bool {
	return true
}
