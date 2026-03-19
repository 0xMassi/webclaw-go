package webclaw

import (
	"context"
	"fmt"
	"time"
)

// ResearchRequest configures an async research job.
type ResearchRequest struct {
	Query         string `json:"query"`
	MaxIterations int    `json:"max_iterations,omitempty"`
	MaxSources    int    `json:"max_sources,omitempty"`
	Topic         string `json:"topic,omitempty"`
	Deep          bool   `json:"deep,omitempty"`
}

// ResearchStartResponse is returned when a research job is started.
type ResearchStartResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ResearchSource represents a source found during research.
type ResearchSource struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// ResearchFinding represents a finding from research.
type ResearchFinding struct {
	Claim     string  `json:"claim"`
	Source    string  `json:"source"`
	Relevance float64 `json:"relevance"`
}

// ResearchStatusResponse contains the status and results of a research job.
type ResearchStatusResponse struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	Query      string            `json:"query"`
	Report     string            `json:"report,omitempty"`
	Sources    []ResearchSource  `json:"sources,omitempty"`
	Findings   []ResearchFinding `json:"findings,omitempty"`
	Iterations int               `json:"iterations,omitempty"`
	ElapsedMs  int               `json:"elapsed_ms,omitempty"`
}

// Research starts an async research job and returns the job ID.
func (c *Client) Research(ctx context.Context, req ResearchRequest) (*ResearchStartResponse, error) {
	var resp ResearchStartResponse
	if err := c.do(ctx, "POST", "/v1/research", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetResearchStatus retrieves the status and results of a research job.
func (c *Client) GetResearchStatus(ctx context.Context, id string) (*ResearchStatusResponse, error) {
	var resp ResearchStatusResponse
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/research/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResearchPollOptions configures polling behavior for WaitForResearch.
type ResearchPollOptions struct {
	Interval time.Duration
	Timeout  time.Duration
}

// WaitForResearch polls a research job until it completes or times out.
func (c *Client) WaitForResearch(ctx context.Context, id string, opts *ResearchPollOptions) (*ResearchStatusResponse, error) {
	interval := 2 * time.Second
	timeout := 5 * time.Minute
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		status, err := c.GetResearchStatus(ctx, id)
		if err != nil {
			return nil, err
		}
		if status.Status == "completed" || status.Status == "failed" {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
