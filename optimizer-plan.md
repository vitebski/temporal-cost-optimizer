# Namespace Workflow Optimizer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement namespace-scoped workflow execution analysis for the optimizer backend by fetching the latest completed workflow history from Temporal and running MVP cost heuristics.

**Architecture:** `internal/httpapi` accepts `namespace` on the workflow analysis endpoint and passes it through `internal/domain.Optimizer`. `internal/temporalcloud` owns all Temporal Cloud and Workflow Service RPC details, including completed-run lookup and paginated history fetch. `internal/optimizer` stays focused on analyzing history events and returning stable `domain.WorkflowAnalysis` responses.

**Tech Stack:** Go 1.23, Temporal Cloud Go SDK, `go.temporal.io/api` workflow service/history protos, standard `testing` package.

---

## File Structure

- Modify `internal/domain/models.go`: change the optimizer interface to accept namespace plus workflow ID.
- Modify `internal/httpapi/router.go`: require `namespace` query parameter for `/workflows/{workflowId}/optimize`.
- Modify `internal/httpapi/router_test.go`: update optimizer fake and add namespace validation tests.
- Modify `internal/config/config.go`: add Temporal workflow service host and history page size configuration.
- Modify `internal/config/config_test.go`: cover the new config defaults and `.env` parsing.
- Modify `.env.example`: document the workflow history connection settings.
- Modify `internal/temporalcloud/client.go`: add workflow service client construction, latest completed run lookup, and history pagination.
- Modify `internal/temporalcloud/client_test.go`: test workflow history request construction and pagination.
- Modify `internal/optimizer/service.go`: implement history-based heuristic analysis.
- Create `internal/optimizer/service_test.go`: cover optimizer heuristics and error propagation.
- Modify `README.md`: update API docs and environment docs after implementation.

---

## Task 1: Domain And HTTP Namespace Plumbing

**Files:**
- Modify: `internal/domain/models.go`
- Modify: `internal/httpapi/router.go`
- Modify: `internal/httpapi/router_test.go`

- [ ] **Step 1: Update the domain optimizer contract**

Change the optimizer interface in `internal/domain/models.go`:

```go
type Optimizer interface {
	AnalyzeWorkflow(ctx context.Context, namespace string, workflowID string) (WorkflowAnalysis, error)
}
```

- [ ] **Step 2: Update the HTTP handler to require namespace**

In `internal/httpapi/router.go`, change `handleWorkflowAnalysis` so it validates `namespace` exactly like workflow type usage does:

```go
func (r *Router) handleWorkflowAnalysis(w http.ResponseWriter, req *http.Request) {
	workflowID, ok := pathBetween(req.URL.Path, "/workflows/", "/optimize")
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "Route not found.")
		return
	}

	namespace := req.URL.Query().Get("namespace")
	if namespace == "" {
		writeError(w, http.StatusBadRequest, "missing_namespace", "Query parameter namespace is required.")
		return
	}

	analysis, err := r.optimizer.AnalyzeWorkflow(req.Context(), namespace, workflowID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}
```

- [ ] **Step 3: Update HTTP tests**

Update `TestRouterReturnsWorkflowAnalysis` to call:

```go
req := httptest.NewRequest(http.MethodGet, "/workflows/wf-123/optimize?namespace=payments-prod", nil)
```

Assert the fake received both values:

```go
if optimizer.namespace != "payments-prod" {
	t.Fatalf("namespace = %q, want payments-prod", optimizer.namespace)
}
if optimizer.workflowID != "wf-123" {
	t.Fatalf("workflow ID = %q, want wf-123", optimizer.workflowID)
}
```

Add a missing namespace test:

```go
func TestRouterRequiresNamespaceForWorkflowAnalysis(t *testing.T) {
	handler := NewRouter(&fakeAnalyzer{}, &fakeOptimizer{})

	req := httptest.NewRequest(http.MethodGet, "/workflows/wf-123/optimize", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "missing_namespace" {
		t.Fatalf("error code = %q, want missing_namespace", body.Error.Code)
	}
}
```

Update the fake optimizer:

```go
type fakeOptimizer struct {
	namespace  string
	workflowID string
	err        error
}

func (f *fakeOptimizer) AnalyzeWorkflow(_ context.Context, namespace string, workflowID string) (domain.WorkflowAnalysis, error) {
	f.namespace = namespace
	f.workflowID = workflowID
	if f.err != nil {
		return domain.WorkflowAnalysis{}, f.err
	}
	return domain.WorkflowAnalysis{
		WorkflowID:    workflowID,
		WorkflowRunID: workflowID + "-run-001",
		Signals: []domain.AnalysisFinding{
			{
				Type:     "large_payload",
				Severity: "high",
				Evidence: "3 events exceed payload threshold",
			},
		},
		Recommendations: []string{
			"Compress large payloads before storing them in workflow state.",
		},
	}, nil
}
```

- [ ] **Step 4: Run focused tests**

Run: `go test ./internal/domain ./internal/httpapi`

Expected: tests pass after updating interface call sites.

---

## Task 2: Configuration For Workflow History Access

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `.env.example`

- [ ] **Step 1: Add config fields**

Extend `TemporalConfig`:

```go
type TemporalConfig struct {
	APIKey        string
	APIHostPort   string
	APIVersion    string
	UsagePageSize int32
}
```

No workflow host field is needed. The backend resolves the namespace workflow endpoint from Cloud Ops namespace metadata at analysis time.

- [ ] **Step 2: Update config tests**

Keep config tests focused on Cloud Ops API settings and usage page size.

- [ ] **Step 3: Update `.env.example`**

Do not add workflow endpoint configuration. It is discovered from Cloud Ops.

- [ ] **Step 4: Run focused tests**

Run: `go test ./internal/config`

Expected: config tests pass.

---

## Task 3: Temporal Cloud Workflow History Client

**Files:**
- Modify: `internal/temporalcloud/client.go`
- Modify: `internal/temporalcloud/client_test.go`

- [ ] **Step 1: Add workflow history types**

In `internal/temporalcloud/client.go`, add:

```go
type WorkflowHistory struct {
	Namespace  string
	WorkflowID string
	RunID      string
	Events     []*history.HistoryEvent
}
```

Import:

```go
common "go.temporal.io/api/common/v1"
enums "go.temporal.io/api/enums/v1"
history "go.temporal.io/api/history/v1"
workflowservice "go.temporal.io/api/workflowservice/v1"
```

- [ ] **Step 2: Add workflow service abstraction**

Add:

```go
type workflowService interface {
	ListWorkflowExecutions(context.Context, *workflowservice.ListWorkflowExecutionsRequest, ...grpc.CallOption) (*workflowservice.ListWorkflowExecutionsResponse, error)
	GetWorkflowExecutionHistory(context.Context, *workflowservice.GetWorkflowExecutionHistoryRequest, ...grpc.CallOption) (*workflowservice.GetWorkflowExecutionHistoryResponse, error)
}
```

Extend `Client`:

```go
type Client struct {
	config          config.TemporalConfig
	usageService    cloudUsageService
	workflowFactory workflowServiceFactory
	closer          closer
}
```

- [ ] **Step 3: Resolve and build the workflow service client**

In `NewClient`, keep the existing Cloud SDK client for usage data and store a workflow service factory:

```go
return newClientForServices(cfg, sdkClient.CloudService(), workflowServiceFactoryFunc(newWorkflowService), sdkClient), nil
```

Implement `newWorkflowService`:

```go
func newWorkflowService(cfg config.TemporalConfig, endpoint string) (workflowService, closer, error) {
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
```

Add local credentials:

```go
type apiKeyCredentials struct {
	apiKey string
}

func (c apiKeyCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + c.apiKey}, nil
}

func (c apiKeyCredentials) RequireTransportSecurity() bool {
	return true
}
```

- [ ] **Step 4: Implement latest completed history lookup**

Add:

```go
func (c *Client) LastCompletedWorkflowHistory(ctx context.Context, namespace string, workflowID string) (WorkflowHistory, error) {
	if c.workflowService == nil {
		return WorkflowHistory{}, domain.ErrNotImplemented
	}
	runID, err := c.lastCompletedRunID(ctx, namespace, workflowID)
	if err != nil {
		return WorkflowHistory{}, err
	}
	events, err := c.workflowHistory(ctx, namespace, workflowID, runID)
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
```

Use visibility query:

```go
func completedWorkflowQuery(workflowID string) string {
	return `WorkflowId = "` + strings.ReplaceAll(strings.ReplaceAll(workflowID, `\`, `\\`), `"`, `\"`) + `" AND ExecutionStatus = "Completed"`
}
```

Page through `ListWorkflowExecutions`, select the completed execution with the newest `CloseTime`, and return its run ID. Treat no completed execution as a normal upstream error:

```go
return "", fmt.Errorf("no completed workflow execution found for namespace %q workflow %q", namespace, workflowID)
```

- [ ] **Step 5: Implement paginated history fetch**

Fetch all history pages:

```go
req := &workflowservice.GetWorkflowExecutionHistoryRequest{
	Namespace: namespace,
	Execution: &common.WorkflowExecution{
		WorkflowId: workflowID,
		RunId:      runID,
	},
	MaximumPageSize: defaultHistoryPageSize,
}
```

Append `resp.GetHistory().GetEvents()` until `NextPageToken` is empty.

- [ ] **Step 6: Close both connections**

Update `Close` to close both clients:

```go
func (c *Client) Close() error {
	if c.workflowCloser != nil {
		if err := c.workflowCloser.Close(); err != nil {
			return err
		}
	}
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}
```

- [ ] **Step 7: Add temporalcloud tests**

Add tests in `internal/temporalcloud/client_test.go`:

```go
func TestLastCompletedWorkflowHistoryListsCompletedExecutionsByNamespace(t *testing.T)
func TestLastCompletedWorkflowHistorySelectsMostRecentCompletedRun(t *testing.T)
func TestLastCompletedWorkflowHistoryFetchesAllHistoryPages(t *testing.T)
func TestLastCompletedWorkflowHistoryReturnsNotImplementedWithoutWorkflowService(t *testing.T)
func TestLastCompletedWorkflowHistoryReturnsErrorWhenNoCompletedRunExists(t *testing.T)
```

Use a fake workflow service that records `ListWorkflowExecutionsRequest` and `GetWorkflowExecutionHistoryRequest`.

- [ ] **Step 8: Run focused tests**

Run: `go test ./internal/temporalcloud`

Expected: temporalcloud tests pass.

---

## Task 4: Optimizer Heuristic Analysis

**Files:**
- Modify: `internal/optimizer/service.go`
- Create: `internal/optimizer/service_test.go`

- [ ] **Step 1: Define the optimizer history dependency**

Replace the current `TemporalHistoryClient` with:

```go
type TemporalHistoryClient interface {
	LastCompletedWorkflowHistory(ctx context.Context, namespace string, workflowID string) (temporalcloud.WorkflowHistory, error)
}
```

Keep:

```go
func NewService(temporal TemporalHistoryClient) *Service {
	return &Service{temporal: temporal}
}
```

- [ ] **Step 2: Implement `AnalyzeWorkflow`**

Use:

```go
func (s *Service) AnalyzeWorkflow(ctx context.Context, namespace string, workflowID string) (domain.WorkflowAnalysis, error) {
	history, err := s.temporal.LastCompletedWorkflowHistory(ctx, namespace, workflowID)
	if err != nil {
		return domain.WorkflowAnalysis{}, err
	}

	findings := analyzeEvents(history.Events)
	return domain.WorkflowAnalysis{
		WorkflowID:      workflowID,
		WorkflowRunID:   history.RunID,
		Signals:         findings,
		Recommendations: recommendationsFor(findings),
	}, nil
}
```

- [ ] **Step 3: Add heuristic constants**

Use MVP thresholds:

```go
const (
	largePayloadThresholdBytes = 64 * 1024
	excessiveSignalThreshold   = 10
	redundantActivityThreshold = 2
	historyBloatThreshold      = 1000
)
```

- [ ] **Step 4: Implement event analysis**

Detect:

- Large payloads from workflow started input, signal input, and activity scheduled input.
- Excessive signals by counting `EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED`.
- Redundant activities by grouping `activity type + payload fingerprint`.
- History bloat by `len(events) > historyBloatThreshold`.

Return findings in stable order:

```go
large_payload
excessive_signals
redundant_activities
history_bloat
```

- [ ] **Step 5: Implement recommendations**

Map finding types:

```go
var recommendationByType = map[string]string{
	"large_payload":        "Compress large payloads before storing them in workflow state.",
	"excessive_signals":    "Batch signals where possible.",
	"redundant_activities": "Deduplicate repeated activities using memoization or cached results.",
	"history_bloat":        "Reduce workflow history size with continue-as-new or by moving large state outside workflow history.",
}
```

Deduplicate recommendations while preserving finding order.

- [ ] **Step 6: Add optimizer tests**

Create tests:

```go
func TestAnalyzeWorkflowDetectsLargePayloads(t *testing.T)
func TestAnalyzeWorkflowDetectsExcessiveSignals(t *testing.T)
func TestAnalyzeWorkflowDetectsRedundantActivities(t *testing.T)
func TestAnalyzeWorkflowDetectsHistoryBloat(t *testing.T)
func TestAnalyzeWorkflowReturnsNoFindingsForSmallHistory(t *testing.T)
func TestAnalyzeWorkflowPropagatesHistoryErrors(t *testing.T)
```

Use real Temporal history protobuf event structs in tests so field access stays honest.

- [ ] **Step 7: Run focused tests**

Run: `go test ./internal/optimizer`

Expected: optimizer tests pass.

---

## Task 5: Docs And Full Verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README environment docs**

Do not add a workflow host environment variable. The backend resolves it from Cloud Ops.

Clarify that workflow analysis requires `namespace`:

```markdown
GET /workflows/{workflowId}/optimize?namespace={namespace}
```

- [ ] **Step 2: Update current implementation status**

Replace the sentence saying workflow execution analysis returns `501` with:

```markdown
`GET /workflows/{workflowId}/optimize?namespace={namespace}` fetches the latest completed run for that workflow ID in the namespace and returns heuristic optimization findings. The backend discovers the namespace workflow endpoint through the Cloud Ops API.
```

- [ ] **Step 3: Run all tests**

Run: `go test ./...`

Expected:

```text
?   	temporal-cost-optimizer/cmd/api	[no test files]
ok  	temporal-cost-optimizer/internal/analyzer
ok  	temporal-cost-optimizer/internal/config
?   	temporal-cost-optimizer/internal/domain	[no test files]
ok  	temporal-cost-optimizer/internal/httpapi
ok  	temporal-cost-optimizer/internal/optimizer
ok  	temporal-cost-optimizer/internal/temporalcloud
```

- [ ] **Step 4: Optional manual API check**

With a running backend and Temporal Cloud API credentials:

```sh
curl "http://localhost:8080/workflows/wf-123/optimize?namespace=payments-prod"
```

Expected response shape:

```json
{
  "workflowId": "wf-123",
  "workflowRunId": "wf-123-run-001",
  "signals": [
    {
      "type": "large_payload",
      "severity": "high",
      "evidence": "3 events exceed payload threshold"
    }
  ],
  "recommendations": [
    "Compress large payloads before storing them in workflow state."
  ]
}
```

---

## Self-Review Notes

- Spec coverage: covers namespace-scoped latest completed execution analysis, large payloads, excessive signals, redundant activities, history bloat, and stable API response shape.
- Scope: touches approved packages plus config/docs needed for runtime wiring.
- Risk: namespace endpoint metadata may be unavailable or lack an API-key gRPC address for some namespaces; surface the upstream error clearly.
- Compatibility: current route path remains unchanged; only the `namespace` query parameter becomes required for analysis.
