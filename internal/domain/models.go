package domain

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("not implemented")

type Analyzer interface {
	TopNamespaces(ctx context.Context, top int) ([]NamespaceSummary, error)
	TopWorkflowTypes(ctx context.Context, namespace string, top int) ([]WorkflowTypeSummary, error)
	WorkflowUsage(ctx context.Context, namespace string, workflowType string) (WorkflowUsage, error)
}

type Optimizer interface {
	AnalyzeWorkflow(ctx context.Context, workflowID string) (WorkflowAnalysis, error)
}

type StorageBreakdown struct {
	Active   StorageUsage `json:"active"`
	Retained StorageUsage `json:"retained"`
}

type StorageUsage struct {
	Usage float64 `json:"usage"`
	Cost  float64 `json:"cost"`
}

type NamespaceSummary struct {
	Namespace     string           `json:"namespace"`
	Rank          int              `json:"rank"`
	UsageScore    float64          `json:"usageScore"`
	EstimatedCost float64          `json:"estimatedCost"`
	Storage       StorageBreakdown `json:"storage"`
	Trend         string           `json:"trend"`
	Incomplete    bool             `json:"incomplete"`
}

type WorkflowTypeSummary struct {
	Namespace     string           `json:"namespace"`
	WorkflowType  string           `json:"workflowType"`
	UsageScore    float64          `json:"usageScore"`
	EstimatedCost float64          `json:"estimatedCost"`
	Storage       StorageBreakdown `json:"storage"`
	Executions    int              `json:"executions"`
	Signals       int              `json:"signals"`
	Activities    int              `json:"activities"`
}

type WorkflowUsage struct {
	WorkflowType string               `json:"workflowType"`
	Namespace    string               `json:"namespace"`
	Summary      WorkflowUsageSummary `json:"summary"`
}

type WorkflowUsageSummary struct {
	Storage          StorageBreakdown `json:"storage"`
	Executions       int              `json:"executions"`
	BillableActions  int              `json:"billableActions"`
	AvgHistoryEvents int              `json:"avgHistoryEvents"`
	P95HistoryEvents int              `json:"p95HistoryEvents"`
}

type WorkflowAnalysis struct {
	WorkflowID      string            `json:"workflowId"`
	WorkflowRunID   string            `json:"workflowRunId"`
	Signals         []AnalysisFinding `json:"signals"`
	Recommendations []string          `json:"recommendations"`
}

type AnalysisFinding struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Evidence string `json:"evidence"`
}
