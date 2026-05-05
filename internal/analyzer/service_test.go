package analyzer

import (
	"context"
	"testing"

	usage "go.temporal.io/cloud-sdk/api/usage/v1"

	"temporal-cost-optimizer/internal/temporalcloud"
)

func TestTopNamespacesAggregatesUsageSummariesByNamespace(t *testing.T) {
	client := &fakeUsageClient{
		pages: []temporalcloud.UsagePage{
			{
				Summaries: []*usage.Summary{
					{
						RecordGroups: []*usage.RecordGroup{
							recordGroup("payments-prod",
								record(usage.RecordType_RECORD_TYPE_ACTIONS, 10),
								record(usage.RecordType_RECORD_TYPE_ACTIVE_STORAGE, 100),
								record(usage.RecordType_RECORD_TYPE_RETAINED_STORAGE, 50),
							),
							recordGroup("search-prod",
								record(usage.RecordType_RECORD_TYPE_ACTIONS, 120),
							),
						},
					},
				},
			},
		},
	}
	service := NewService(client)

	items, err := service.TopNamespaces(context.Background(), 1)
	if err != nil {
		t.Fatalf("TopNamespaces returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("items length = %d, want 1", len(items))
	}
	if items[0].Namespace != "payments-prod" {
		t.Fatalf("top namespace = %q, want payments-prod", items[0].Namespace)
	}
	if items[0].Rank != 1 {
		t.Fatalf("rank = %d, want 1", items[0].Rank)
	}
	if items[0].UsageScore != 160 {
		t.Fatalf("usage score = %v, want 160", items[0].UsageScore)
	}
	if items[0].Storage.Active.Usage != 100 {
		t.Fatalf("active storage usage = %v, want 100", items[0].Storage.Active.Usage)
	}
	if client.queries[0].PageSize != 0 {
		t.Fatalf("usage query page size = %d, want 0 so client config can apply", client.queries[0].PageSize)
	}
}

func TestTopNamespacesFetchesAllUsagePages(t *testing.T) {
	client := &fakeUsageClient{
		pages: []temporalcloud.UsagePage{
			{
				Summaries: []*usage.Summary{
					{
						RecordGroups: []*usage.RecordGroup{
							recordGroup("payments-prod", record(usage.RecordType_RECORD_TYPE_ACTIONS, 10)),
						},
					},
				},
				NextPageToken: "page-2",
			},
			{
				Summaries: []*usage.Summary{
					{
						RecordGroups: []*usage.RecordGroup{
							recordGroup("payments-prod", record(usage.RecordType_RECORD_TYPE_ACTIONS, 15)),
						},
					},
				},
			},
		},
	}
	service := NewService(client)

	items, err := service.TopNamespaces(context.Background(), 5)
	if err != nil {
		t.Fatalf("TopNamespaces returned error: %v", err)
	}

	if len(client.queries) != 2 {
		t.Fatalf("usage calls = %d, want 2", len(client.queries))
	}
	if client.queries[1].PageToken != "page-2" {
		t.Fatalf("second page token = %q, want page-2", client.queries[1].PageToken)
	}
	if items[0].UsageScore != 25 {
		t.Fatalf("usage score = %v, want 25", items[0].UsageScore)
	}
}

func TestTopNamespacesMarksIncompleteUsage(t *testing.T) {
	client := &fakeUsageClient{
		pages: []temporalcloud.UsagePage{
			{
				Summaries: []*usage.Summary{
					{
						Incomplete: true,
						RecordGroups: []*usage.RecordGroup{
							recordGroup("payments-prod", record(usage.RecordType_RECORD_TYPE_ACTIONS, 10)),
						},
					},
				},
			},
		},
	}
	service := NewService(client)

	items, err := service.TopNamespaces(context.Background(), 5)
	if err != nil {
		t.Fatalf("TopNamespaces returned error: %v", err)
	}

	if !items[0].Incomplete {
		t.Fatal("namespace summary incomplete = false, want true")
	}
}

type fakeUsageClient struct {
	queries []temporalcloud.UsageQuery
	pages   []temporalcloud.UsagePage
	err     error
}

func (f *fakeUsageClient) GetUsage(_ context.Context, query temporalcloud.UsageQuery) (temporalcloud.UsagePage, error) {
	f.queries = append(f.queries, query)
	if f.err != nil {
		return temporalcloud.UsagePage{}, f.err
	}
	if len(f.pages) == 0 {
		return temporalcloud.UsagePage{}, nil
	}
	page := f.pages[0]
	f.pages = f.pages[1:]
	return page, nil
}

func recordGroup(namespace string, records ...*usage.Record) *usage.RecordGroup {
	return &usage.RecordGroup{
		GroupBys: []*usage.GroupBy{
			{
				Key:   usage.GroupByKey_GROUP_BY_KEY_NAMESPACE,
				Value: namespace,
			},
		},
		Records: records,
	}
}

func record(recordType usage.RecordType, value float64) *usage.Record {
	return &usage.Record{
		Type:  recordType,
		Value: value,
	}
}
