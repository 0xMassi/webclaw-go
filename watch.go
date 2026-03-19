package webclaw

import (
	"context"
	"fmt"
)

// WatchCreateRequest configures a new URL watch monitor.
type WatchCreateRequest struct {
	URL             string `json:"url"`
	Name            string `json:"name,omitempty"`
	IntervalMinutes int    `json:"interval_minutes,omitempty"`
	WebhookURL      string `json:"webhook_url,omitempty"`
}

// WatchEntry represents a single watch monitor.
type WatchEntry struct {
	ID              string `json:"id"`
	URL             string `json:"url"`
	Name            string `json:"name,omitempty"`
	IntervalMinutes int    `json:"interval_minutes"`
	Active          bool   `json:"active"`
	WebhookURL      string `json:"webhook_url,omitempty"`
	LastCheckedAt   string `json:"last_checked_at,omitempty"`
	LastChangedAt   string `json:"last_changed_at,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

// WatchSnapshot represents a single point-in-time check of a watched URL.
type WatchSnapshot struct {
	ID             string `json:"id"`
	ContentHash    string `json:"content_hash"`
	WordCount      int    `json:"word_count"`
	Status         string `json:"status"`
	Title          string `json:"title,omitempty"`
	DiffSummary    string `json:"diff_summary,omitempty"`
	WordCountDelta int    `json:"word_count_delta"`
	LinksAdded     int    `json:"links_added"`
	LinksRemoved   int    `json:"links_removed"`
	CheckedAt      string `json:"checked_at"`
}

// WatchDetail is a watch entry with its recent snapshots.
type WatchDetail struct {
	WatchEntry
	Snapshots []WatchSnapshot `json:"snapshots,omitempty"`
}

// WatchListResponse is the paginated list of watches.
type WatchListResponse struct {
	Watches []WatchEntry `json:"watches"`
}

// WatchCheckResponse is returned when a manual check is triggered.
type WatchCheckResponse struct {
	Status string `json:"status"`
}

// WatchCreate creates a new URL watch monitor.
func (c *Client) WatchCreate(ctx context.Context, req *WatchCreateRequest) (*WatchEntry, error) {
	var resp WatchEntry
	if err := c.do(ctx, "POST", "/v1/watch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WatchList returns the authenticated user's watches with pagination.
func (c *Client) WatchList(ctx context.Context, limit, offset int) (*WatchListResponse, error) {
	path := fmt.Sprintf("/v1/watch?limit=%d&offset=%d", limit, offset)
	var resp WatchListResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WatchGet retrieves a single watch by ID, including recent snapshots.
func (c *Client) WatchGet(ctx context.Context, id string) (*WatchDetail, error) {
	var resp WatchDetail
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/watch/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WatchDelete removes a watch and all its snapshots.
func (c *Client) WatchDelete(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/v1/watch/%s", id), nil, nil)
}

// WatchCheck triggers an immediate check of a watched URL.
func (c *Client) WatchCheck(ctx context.Context, id string) (*WatchCheckResponse, error) {
	var resp WatchCheckResponse
	if err := c.do(ctx, "POST", fmt.Sprintf("/v1/watch/%s/check", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
