package temporalcloud

import (
	"context"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	usage "go.temporal.io/cloud-sdk/api/usage/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"temporal-cost-optimizer/internal/config"
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

type cloudUsageService interface {
	GetUsage(context.Context, *cloudservice.GetUsageRequest, ...grpc.CallOption) (*cloudservice.GetUsageResponse, error)
}

type closer interface {
	Close() error
}

type Client struct {
	config       config.TemporalConfig
	usageService cloudUsageService
	closer       closer
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

	return newClientForUsageService(cfg, sdkClient.CloudService(), sdkClient), nil
}

func newClientForUsageService(cfg config.TemporalConfig, usageService cloudUsageService, closer closer) *Client {
	return &Client{
		config:       cfg,
		usageService: usageService,
		closer:       closer,
	}
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

	resp, err := c.usageService.GetUsage(ctx, req)
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

func (c *Client) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}
