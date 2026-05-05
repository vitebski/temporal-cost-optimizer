package temporalcloud

import "temporal-cost-optimizer/internal/config"

type Client struct {
	config config.TemporalConfig
}

func NewClient(cfg config.TemporalConfig) *Client {
	return &Client{config: cfg}
}

func (c *Client) Config() config.TemporalConfig {
	return c.config
}
