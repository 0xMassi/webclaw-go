package webclaw

import (
	"context"
	"fmt"
	"time"
)

// Crawl starts an async crawl job and returns immediately with the job ID.
func (c *Client) Crawl(ctx context.Context, req *CrawlRequest) (*CrawlStartResponse, error) {
	var resp CrawlStartResponse
	if err := c.do(ctx, "POST", "/v1/crawl", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCrawl polls the status of a crawl job by ID.
func (c *Client) GetCrawl(ctx context.Context, id string) (*CrawlStatusResponse, error) {
	var resp CrawlStatusResponse
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/crawl/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WaitForCompletion polls a crawl job until it reaches a terminal state
// (completed or failed) or the context is cancelled. The interval controls
// how often it polls; pass 0 for a 2-second default.
func (c *Client) WaitForCompletion(ctx context.Context, id string, interval time.Duration) (*CrawlStatusResponse, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		resp, err := c.GetCrawl(ctx, id)
		if err != nil {
			return nil, err
		}
		if resp.Status == CrawlStatusCompleted || resp.Status == CrawlStatusFailed {
			return resp, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// next poll
		}
	}
}
