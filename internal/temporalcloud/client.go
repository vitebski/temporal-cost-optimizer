package temporalcloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
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
	"google.golang.org/grpc/metadata"
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

type multiCloser []closer

func (m multiCloser) Close() error {
	var firstErr error
	for _, closer := range m {
		if closer == nil {
			continue
		}
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type Client struct {
	config                config.TemporalConfig
	usageCloudService     cloudService
	namespaceCloudService cloudService
	workflowFactory       workflowServiceFactory
	closer                closer
}

func NewClient(cfg config.TemporalConfig) (*Client, error) {
	log.Printf(
		"temporal_client_config usage_key_present=%t usage_key_fingerprint=%q namespace_key_present=%t namespace_key_fingerprint=%q api_host_port=%q api_version=%q",
		cfg.UsageAPIKey != "",
		keyFingerprint(cfg.UsageAPIKey),
		cfg.NamespaceAPIKey != "",
		keyFingerprint(cfg.NamespaceAPIKey),
		cfg.APIHostPort,
		cfg.APIVersion,
	)

	usageSDKClient, err := cloudclient.New(cloudclient.Options{
		APIKey:     cfg.UsageAPIKey,
		HostPort:   cfg.APIHostPort,
		APIVersion: cfg.APIVersion,
		UserAgent:  "temporal-cost-optimizer",
	})
	if err != nil {
		return nil, err
	}

	namespaceSDKClient, err := cloudclient.New(cloudclient.Options{
		APIKey:     cfg.NamespaceAPIKey,
		HostPort:   cfg.APIHostPort,
		APIVersion: cfg.APIVersion,
		UserAgent:  "temporal-cost-optimizer",
	})
	if err != nil {
		_ = usageSDKClient.Close()
		return nil, err
	}

	return newClientForServices(
		cfg,
		usageSDKClient.CloudService(),
		namespaceSDKClient.CloudService(),
		workflowServiceFactoryFunc(newWorkflowService),
		multiCloser{usageSDKClient, namespaceSDKClient},
	), nil
}

func newClientForUsageService(cfg config.TemporalConfig, cloudService cloudService, closer closer) *Client {
	return newClientForServices(cfg, cloudService, nil, workflowServiceFactoryFunc(newWorkflowService), closer)
}

func newClientForServices(cfg config.TemporalConfig, usageCloudService cloudService, namespaceCloudService cloudService, workflowFactory workflowServiceFactory, closer closer) *Client {
	return &Client{
		config:                cfg,
		usageCloudService:     usageCloudService,
		namespaceCloudService: namespaceCloudService,
		workflowFactory:       workflowFactory,
		closer:                closer,
	}
}

func newWorkflowService(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error) {
	log.Printf(
		"temporal_workflow_credentials role=namespace endpoint=%q key_present=%t key_fingerprint=%q",
		endpoint,
		cfg.NamespaceAPIKey != "",
		keyFingerprint(cfg.NamespaceAPIKey),
	)

	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(apiKeyCredentials{apiKey: cfg.NamespaceAPIKey}),
		grpc.WithUserAgent("temporal-cost-optimizer"),
	)
	if err != nil {
		return nil, nil, err
	}

	return workflowservice.NewWorkflowServiceClient(conn), conn, nil
}

func keyFingerprint(key string) string {
	if key == "" {
		return "missing"
	}
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:12]
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

	resp, err := c.usageCloudService.GetUsage(ctx, req)
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
	if c.namespaceCloudService == nil || c.workflowFactory == nil {
		return WorkflowHistory{}, domain.ErrNotImplemented
	}

	startedAt := time.Now()
	log.Printf("temporal_workflow_history_start namespace=%q workflow_id=%q", namespace, workflowID)
	endpoint, err := c.workflowEndpoint(ctx, namespace)
	if err != nil {
		log.Printf("temporal_workflow_history_error namespace=%q workflow_id=%q step=resolve_endpoint duration=%s err=%q", namespace, workflowID, time.Since(startedAt), err.Error())
		return WorkflowHistory{}, err
	}
	log.Printf("temporal_workflow_endpoint_resolved namespace=%q workflow_id=%q endpoint=%q", namespace, workflowID, endpoint)

	workflowService, workflowCloser, err := c.workflowFactory.NewWorkflowService(c.config, endpoint)
	if err != nil {
		log.Printf("temporal_workflow_connect_error namespace=%q workflow_id=%q endpoint=%q duration=%s err=%q", namespace, workflowID, endpoint, time.Since(startedAt), err.Error())
		return WorkflowHistory{}, fmt.Errorf("connect workflow service endpoint %q: %w", endpoint, err)
	}
	log.Printf("temporal_workflow_connected namespace=%q workflow_id=%q endpoint=%q", namespace, workflowID, endpoint)
	if workflowCloser != nil {
		defer workflowCloser.Close()
	}

	workflowCtx := metadata.AppendToOutgoingContext(ctx, "temporal-namespace", namespace)

	runID, err := c.lastCompletedRunID(workflowCtx, workflowService, namespace, workflowID)
	if err != nil {
		log.Printf("temporal_workflow_history_error namespace=%q workflow_id=%q step=list_completed_runs duration=%s err=%q", namespace, workflowID, time.Since(startedAt), err.Error())
		return WorkflowHistory{}, err
	}
	log.Printf("temporal_workflow_run_selected namespace=%q workflow_id=%q run_id=%q", namespace, workflowID, runID)

	events, err := c.workflowHistory(workflowCtx, workflowService, namespace, workflowID, runID)
	if err != nil {
		log.Printf("temporal_workflow_history_error namespace=%q workflow_id=%q run_id=%q step=get_history duration=%s err=%q", namespace, workflowID, runID, time.Since(startedAt), err.Error())
		return WorkflowHistory{}, err
	}
	log.Printf("temporal_workflow_history_complete namespace=%q workflow_id=%q run_id=%q events=%d duration=%s", namespace, workflowID, runID, len(events), time.Since(startedAt))

	return WorkflowHistory{
		Namespace:  namespace,
		WorkflowID: workflowID,
		RunID:      runID,
		Events:     events,
	}, nil
}

func (c *Client) workflowEndpoint(ctx context.Context, namespace string) (string, error) {
	log.Printf("temporal_namespace_lookup_start namespace=%q", namespace)
	resp, err := c.namespaceCloudService.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: namespace})
	if err != nil {
		log.Printf("temporal_namespace_lookup_error namespace=%q err=%q", namespace, err.Error())
		return "", fmt.Errorf("get namespace %q from Cloud Ops: %w", namespace, err)
	}

	endpoint := resp.GetNamespace().GetEndpoints().GetGrpcAddress()
	if endpoint == "" {
		log.Printf("temporal_namespace_lookup_error namespace=%q err=%q", namespace, "missing API-key gRPC endpoint")
		return "", fmt.Errorf("namespace %q has no API-key gRPC endpoint: %w", namespace, domain.ErrNotImplemented)
	}

	log.Printf("temporal_namespace_lookup_complete namespace=%q endpoint=%q", namespace, endpoint)
	return endpoint, nil
}

func (c *Client) lastCompletedRunID(ctx context.Context, workflowService workflowService, namespace string, workflowID string) (string, error) {
	var selectedRunID string
	var selectedCloseTime time.Time
	nextPageToken := []byte(nil)

	for {
		log.Printf("temporal_workflow_list_start namespace=%q workflow_id=%q page_size=%d has_page_token=%t", namespace, workflowID, defaultWorkflowListPageSize, len(nextPageToken) > 0)
		resp, err := workflowService.ListWorkflowExecutions(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      defaultWorkflowListPageSize,
			NextPageToken: nextPageToken,
			Query:         completedWorkflowQuery(workflowID),
		})
		if err != nil {
			log.Printf("temporal_workflow_list_error namespace=%q workflow_id=%q err=%q", namespace, workflowID, err.Error())
			return "", fmt.Errorf("list completed workflow executions for namespace %q workflow %q: %w", namespace, workflowID, err)
		}
		log.Printf("temporal_workflow_list_complete namespace=%q workflow_id=%q executions=%d has_next_page=%t", namespace, workflowID, len(resp.GetExecutions()), len(resp.GetNextPageToken()) > 0)

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
		log.Printf("temporal_workflow_run_not_found namespace=%q workflow_id=%q", namespace, workflowID)
		return "", fmt.Errorf("no completed workflow execution found for namespace %q workflow %q", namespace, workflowID)
	}

	return selectedRunID, nil
}

func (c *Client) workflowHistory(ctx context.Context, workflowService workflowService, namespace string, workflowID string, runID string) ([]*history.HistoryEvent, error) {
	var events []*history.HistoryEvent
	nextPageToken := []byte(nil)

	for {
		log.Printf("temporal_workflow_history_page_start namespace=%q workflow_id=%q run_id=%q page_size=%d has_page_token=%t", namespace, workflowID, runID, defaultHistoryPageSize, len(nextPageToken) > 0)
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
			log.Printf("temporal_workflow_history_page_error namespace=%q workflow_id=%q run_id=%q err=%q", namespace, workflowID, runID, err.Error())
			return nil, fmt.Errorf("get workflow history for namespace %q workflow %q run %q: %w", namespace, workflowID, runID, err)
		}

		events = append(events, resp.GetHistory().GetEvents()...)
		log.Printf("temporal_workflow_history_page_complete namespace=%q workflow_id=%q run_id=%q events=%d total_events=%d has_next_page=%t", namespace, workflowID, runID, len(resp.GetHistory().GetEvents()), len(events), len(resp.GetNextPageToken()) > 0)
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
