package webclaw

import (
	"context"
	"fmt"
	"time"
)

// LeadBatchRequest configures an async multi-URL lead enrichment job.
// URLs must hold 1..25 company website URLs; the server validates and
// dedupes them before starting the job.
type LeadBatchRequest struct {
	URLs []string `json:"urls"`
	// NoCache bypasses the cache and forces a fresh enrichment for every URL.
	NoCache bool `json:"no_cache,omitempty"`
}

// LeadBatchStartResponse is returned immediately when a lead batch job is
// started; the job itself runs asynchronously. Poll GetLeadBatch with ID to
// track progress.
type LeadBatchStartResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Total  int    `json:"total"`
	// CreditsPerURL is the credit cost billed per successful lead (100).
	CreditsPerURL int `json:"credits_per_url"`
}

// LeadBatchResult is the outcome for a single URL in a lead batch job. It is a
// tagged union: Status is "success" or "error". On "success" Domain, Lead, and
// Cache are populated and Error is empty; on "error" only Error is set.
type LeadBatchResult struct {
	URL    string `json:"url"`
	Status string `json:"status"`
	// Domain, Lead, and Cache are set when Status == "success".
	Domain string   `json:"domain,omitempty"`
	Lead   LeadData `json:"lead,omitempty"`
	Cache  string   `json:"cache,omitempty"`
	// Error is set when Status == "error".
	Error string `json:"error,omitempty"`
}

// LeadBatchResponse contains the current state and results of a lead batch job.
// Only successful leads are billed (100 credits each); error results are free.
type LeadBatchResponse struct {
	ID             string            `json:"id"`
	Status         string            `json:"status"`
	Total          int               `json:"total"`
	Completed      int               `json:"completed"`
	Succeeded      int               `json:"succeeded"`
	CreditsCharged int               `json:"credits_charged"`
	Results        []LeadBatchResult `json:"results,omitempty"`
	// Error carries a job-level failure message; empty (JSON null) otherwise.
	Error string `json:"error,omitempty"`
	// CreatedAt is the ISO-8601 timestamp when the job was created.
	CreatedAt string `json:"created_at,omitempty"`
}

// LeadBatch starts an async multi-URL lead enrichment job and returns
// immediately with the job ID. Each URL is enriched into a structured sales
// lead (see Lead); poll GetLeadBatch until Status is "completed" or "failed".
//
// Billing: 100 credits per successful lead. Error results are not billed.
func (c *Client) LeadBatch(ctx context.Context, req *LeadBatchRequest) (*LeadBatchStartResponse, error) {
	var resp LeadBatchStartResponse
	if err := c.do(ctx, "POST", "/v1/lead/batch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLeadBatch polls the status and results of a lead batch job by ID.
func (c *Client) GetLeadBatch(ctx context.Context, id string) (*LeadBatchResponse, error) {
	var resp LeadBatchResponse
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/lead/batch/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LeadBatchPollOptions configures polling behavior for WaitForLeadBatch.
type LeadBatchPollOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}

// WaitForLeadBatch polls a lead batch job until it reaches a terminal state
// (completed or failed) or the context is cancelled. Pass nil for defaults
// (2s interval, no timeout beyond the parent context). Set opts.Timeout to
// bound the wait; otherwise the caller's context controls deadlines.
func (c *Client) WaitForLeadBatch(ctx context.Context, id string, opts *LeadBatchPollOptions) (*LeadBatchResponse, error) {
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
		resp, err := c.GetLeadBatch(ctx, id)
		if err != nil {
			return nil, err
		}
		if resp.Status == "completed" || resp.Status == "failed" {
			return resp, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
