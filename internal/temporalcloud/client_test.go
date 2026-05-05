package temporalcloud

import (
	"context"
	"testing"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	usage "go.temporal.io/cloud-sdk/api/usage/v1"
	"google.golang.org/grpc"

	"temporal-cost-optimizer/internal/config"
)

func TestGetUsageBuildsCloudSDKRequest(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	fake := &fakeCloudUsageService{
		response: &cloudservice.GetUsageResponse{
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

	if fake.request.GetStartTimeInclusive().AsTime() != start {
		t.Fatalf("start time = %s, want %s", fake.request.GetStartTimeInclusive().AsTime(), start)
	}
	if fake.request.GetEndTimeExclusive().AsTime() != end {
		t.Fatalf("end time = %s, want %s", fake.request.GetEndTimeExclusive().AsTime(), end)
	}
	if fake.request.GetPageSize() != 25 {
		t.Fatalf("page size = %d, want 25", fake.request.GetPageSize())
	}
	if fake.request.GetPageToken() != "next-token" {
		t.Fatalf("page token = %q, want next-token", fake.request.GetPageToken())
	}
	if len(page.Summaries) != 1 {
		t.Fatalf("summaries length = %d, want 1", len(page.Summaries))
	}
}

func TestGetUsageUsesConfiguredPageSize(t *testing.T) {
	fake := &fakeCloudUsageService{response: &cloudservice.GetUsageResponse{}}
	client := newClientForUsageService(config.TemporalConfig{UsagePageSize: 300}, fake, nil)

	_, err := client.GetUsage(context.Background(), UsageQuery{})
	if err != nil {
		t.Fatalf("GetUsage returned error: %v", err)
	}

	if fake.request.GetPageSize() != 300 {
		t.Fatalf("page size = %d, want configured default 300", fake.request.GetPageSize())
	}
}

type fakeCloudUsageService struct {
	request  *cloudservice.GetUsageRequest
	response *cloudservice.GetUsageResponse
}

func (f *fakeCloudUsageService) GetUsage(ctx context.Context, req *cloudservice.GetUsageRequest, _ ...grpc.CallOption) (*cloudservice.GetUsageResponse, error) {
	f.request = req
	return f.response, nil
}
