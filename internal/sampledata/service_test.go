package sampledata

import (
	"context"
	"testing"

	"temporal-cost-optimizer/internal/domain"
)

func TestServiceImplementsDomainInterfaces(t *testing.T) {
	var analyzer domain.Analyzer = NewService()
	var optimizer domain.Optimizer = NewService()

	if analyzer == nil {
		t.Fatal("analyzer = nil")
	}
	if optimizer == nil {
		t.Fatal("optimizer = nil")
	}
}

func TestTopNamespacesReturnsRequestedLimit(t *testing.T) {
	service := NewService()

	items, err := service.TopNamespaces(context.Background(), 3)
	if err != nil {
		t.Fatalf("TopNamespaces returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("items length = %d, want 3", len(items))
	}
	for i, item := range items {
		if item.Rank != i+1 {
			t.Fatalf("rank for item %d = %d, want %d", i, item.Rank, i+1)
		}
		if item.Namespace == "" {
			t.Fatalf("namespace for item %d is empty", i)
		}
		if item.UsageScore <= 0 {
			t.Fatalf("usage score for item %d = %f, want positive", i, item.UsageScore)
		}
	}
}

func TestTopWorkflowTypesUsesNamespaceAndLimit(t *testing.T) {
	service := NewService()

	items, err := service.TopWorkflowTypes(context.Background(), "payments-prod", 2)
	if err != nil {
		t.Fatalf("TopWorkflowTypes returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items length = %d, want 2", len(items))
	}
	for _, item := range items {
		if item.Namespace != "payments-prod" {
			t.Fatalf("namespace = %q, want payments-prod", item.Namespace)
		}
		if item.WorkflowType == "" {
			t.Fatal("workflow type is empty")
		}
	}
}

func TestWorkflowUsageUsesRequestInputs(t *testing.T) {
	service := NewService()

	usage, err := service.WorkflowUsage(context.Background(), "payments-prod", "ChargeCardWorkflow")
	if err != nil {
		t.Fatalf("WorkflowUsage returned error: %v", err)
	}
	if usage.Namespace != "payments-prod" {
		t.Fatalf("namespace = %q, want payments-prod", usage.Namespace)
	}
	if usage.WorkflowType != "ChargeCardWorkflow" {
		t.Fatalf("workflow type = %q, want ChargeCardWorkflow", usage.WorkflowType)
	}
	if usage.Summary.Executions <= 0 {
		t.Fatalf("executions = %d, want positive", usage.Summary.Executions)
	}
}

func TestAnalyzeWorkflowUsesRequestInputs(t *testing.T) {
	service := NewService()

	analysis, err := service.AnalyzeWorkflow(context.Background(), "payments-prod", "wf-123")
	if err != nil {
		t.Fatalf("AnalyzeWorkflow returned error: %v", err)
	}
	if analysis.WorkflowID != "wf-123" {
		t.Fatalf("workflow ID = %q, want wf-123", analysis.WorkflowID)
	}
	if analysis.WorkflowRunID == "" {
		t.Fatal("workflow run ID is empty")
	}
	if len(analysis.Signals) == 0 {
		t.Fatal("signals length = 0, want sample findings")
	}
	if len(analysis.Recommendations) == 0 {
		t.Fatal("recommendations length = 0, want sample recommendations")
	}
}
