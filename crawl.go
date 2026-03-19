package webclaw

import (
	"context"
	"fmt"
	"time"
)

// CrawlPollOptions configures polling behavior for WaitForCompletion.
type CrawlPollOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}

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
// (completed or failed) or the context is cancelled. Pass nil for defaults
// (2s interval, no timeout beyond the parent context).
func (c *Client) WaitForCompletion(ctx context.Context, id string, opts *CrawlPollOptions) (*CrawlStatusResponse, error) {
	interval := 2 * time.Second
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}
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
