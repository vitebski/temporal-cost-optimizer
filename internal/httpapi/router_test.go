package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"temporal-cost-optimizer/internal/domain"
)

func TestRouterReturnsTopNamespaces(t *testing.T) {
	analyzer := &fakeAnalyzer{}
	handler := NewRouter(analyzer, &fakeOptimizer{})

	req := httptest.NewRequest(http.MethodGet, "/namespaces?top=3", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if analyzer.topNamespacesTop != 3 {
		t.Fatalf("top argument = %d, want 3", analyzer.topNamespacesTop)
	}

	var body struct {
		Items []domain.NamespaceSummary `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := body.Items[0].Namespace; got != "payments-prod" {
		t.Fatalf("namespace = %q, want payments-prod", got)
	}
}

func TestRouterReturnsTopWorkflowTypesForNamespace(t *testing.T) {
	analyzer := &fakeAnalyzer{}
	handler := NewRouter(analyzer, &fakeOptimizer{})

	req := httptest.NewRequest(http.MethodGet, "/namespaces/payments-prod/workflow-types?top=2", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if analyzer.workflowTypesNamespace != "payments-prod" {
		t.Fatalf("namespace argument = %q, want payments-prod", analyzer.workflowTypesNamespace)
	}
	if analyzer.workflowTypesTop != 2 {
		t.Fatalf("top argument = %d, want 2", analyzer.workflowTypesTop)
	}

	var body struct {
		Namespace string                       `json:"namespace"`
		Items     []domain.WorkflowTypeSummary `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Namespace != "payments-prod" {
		t.Fatalf("namespace in response = %q, want payments-prod", body.Namespace)
	}
	if got := body.Items[0].WorkflowType; got != "ChargeCardWorkflow" {
		t.Fatalf("workflow type = %q, want ChargeCardWorkflow", got)
	}
}

func TestRouterReturnsWorkflowUsage(t *testing.T) {
	analyzer := &fakeAnalyzer{}
	handler := NewRouter(analyzer, &fakeOptimizer{})

	req := httptest.NewRequest(http.MethodGet, "/workflow-types/ChargeCardWorkflow/usage?namespace=payments-prod", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if analyzer.usageNamespace != "payments-prod" {
		t.Fatalf("namespace argument = %q, want payments-prod", analyzer.usageNamespace)
	}
	if analyzer.usageWorkflowType != "ChargeCardWorkflow" {
		t.Fatalf("workflow type argument = %q, want ChargeCardWorkflow", analyzer.usageWorkflowType)
	}

	var body domain.WorkflowUsage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Summary.BillableActions != 9100 {
		t.Fatalf("billable actions = %d, want 9100", body.Summary.BillableActions)
	}
}

func TestRouterReturnsWorkflowAnalysis(t *testing.T) {
	optimizer := &fakeOptimizer{}
	handler := NewRouter(&fakeAnalyzer{}, optimizer)

	req := httptest.NewRequest(http.MethodGet, "/workflows/wf-123/analyze", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if optimizer.workflowID != "wf-123" {
		t.Fatalf("workflow ID = %q, want wf-123", optimizer.workflowID)
	}

	var body domain.WorkflowAnalysis
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.WorkflowID != "wf-123" {
		t.Fatalf("workflow ID in response = %q, want wf-123", body.WorkflowID)
	}
}

func TestRouterMapsNotImplementedTo501(t *testing.T) {
	analyzer := &fakeAnalyzer{err: domain.ErrNotImplemented}
	handler := NewRouter(analyzer, &fakeOptimizer{})

	req := httptest.NewRequest(http.MethodGet, "/namespaces", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "not_implemented" {
		t.Fatalf("error code = %q, want not_implemented", body.Error.Code)
	}
}

func TestRouterMapsOptimizerNotImplementedTo501(t *testing.T) {
	optimizer := &fakeOptimizer{err: domain.ErrNotImplemented}
	handler := NewRouter(&fakeAnalyzer{}, optimizer)

	req := httptest.NewRequest(http.MethodGet, "/workflows/wf-123/analyze", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "not_implemented" {
		t.Fatalf("error code = %q, want not_implemented", body.Error.Code)
	}
}

type fakeAnalyzer struct {
	topNamespacesTop       int
	workflowTypesNamespace string
	workflowTypesTop       int
	usageNamespace         string
	usageWorkflowType      string
	err                    error
}

func (f *fakeAnalyzer) TopNamespaces(_ context.Context, top int) ([]domain.NamespaceSummary, error) {
	f.topNamespacesTop = top
	if f.err != nil {
		return nil, f.err
	}
	return []domain.NamespaceSummary{
		{
			Namespace:     "payments-prod",
			Rank:          1,
			UsageScore:    98342,
			EstimatedCost: 124.55,
			Trend:         "up",
		},
	}, nil
}

func (f *fakeAnalyzer) TopWorkflowTypes(_ context.Context, namespace string, top int) ([]domain.WorkflowTypeSummary, error) {
	f.workflowTypesNamespace = namespace
	f.workflowTypesTop = top
	if f.err != nil {
		return nil, f.err
	}
	return []domain.WorkflowTypeSummary{
		{
			Namespace:     "payments-prod",
			WorkflowType:  "ChargeCardWorkflow",
			UsageScore:    44100,
			EstimatedCost: 54.2,
			Signals:       220,
			Activities:    910,
		},
	}, nil
}

func (f *fakeAnalyzer) WorkflowUsage(_ context.Context, namespace string, workflowType string) (domain.WorkflowUsage, error) {
	f.usageNamespace = namespace
	f.usageWorkflowType = workflowType
	if f.err != nil {
		return domain.WorkflowUsage{}, f.err
	}
	return domain.WorkflowUsage{
		Namespace:    "payments-prod",
		WorkflowType: "ChargeCardWorkflow",
		Summary: domain.WorkflowUsageSummary{
			Executions:       182,
			BillableActions:  9100,
			AvgHistoryEvents: 144,
			P95HistoryEvents: 302,
		},
	}, nil
}

type fakeOptimizer struct {
	workflowID string
	err        error
}

func (f *fakeOptimizer) AnalyzeWorkflow(_ context.Context, workflowID string) (domain.WorkflowAnalysis, error) {
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

var _ domain.Analyzer = (*fakeAnalyzer)(nil)
var _ domain.Optimizer = (*fakeOptimizer)(nil)
