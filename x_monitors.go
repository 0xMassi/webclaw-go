package webclaw

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// XMonitorKind classifies what an X (Twitter) monitor polls.
type XMonitorKind string

const (
	// XMonitorProfile watches a single account's timeline (target = handle).
	XMonitorProfile XMonitorKind = "profile"
	// XMonitorSearch watches a search query (target = query string).
	XMonitorSearch XMonitorKind = "search"
	// XMonitorList watches an X list (target = list id).
	XMonitorList XMonitorKind = "list"
	// XMonitorReplies watches replies to a tweet (target = tweet id).
	XMonitorReplies XMonitorKind = "replies"
)

// XMonitorCreateRequest configures a new X (Twitter) monitor. The monitor
// polls X on an interval and fires WebhookURL when new matching tweets appear.
//
// This is the X analog of WatchCreateRequest. It is a paid-only feature: the
// API returns 403 for free or lapsed accounts. Each check (automated or
// manual via XMonitorCheck) costs one credit, and a user may hold at most 50
// monitors.
//
// The pointer-typed booleans (IncludeRetweets, IncludeReplies, IncludeQuotes)
// default to true server-side. Leave them nil to accept the default, or set a
// value explicitly (including false) to override -- a plain bool could not
// distinguish "unset" from an intentional false.
type XMonitorCreateRequest struct {
	// Kind is one of profile, search, list, or replies. Required.
	Kind XMonitorKind `json:"kind"`
	// Target is a handle (a leading @ is stripped), search query, list id, or
	// tweet id, interpreted according to Kind. Required.
	Target string `json:"target"`
	// Name is an optional human-readable label.
	Name string `json:"name,omitempty"`
	// IntervalMinutes is the poll cadence. Defaults to 15, clamped to 2..10080.
	IntervalMinutes int `json:"interval_minutes,omitempty"`
	// WebhookURL is fired on new matches. Discord and Slack URLs receive native
	// embed/text formatting; any other URL receives the generic JSON payload.
	WebhookURL string `json:"webhook_url,omitempty"`
	// IncludeRetweets toggles matching retweets. Defaults to true.
	IncludeRetweets *bool `json:"include_retweets,omitempty"`
	// IncludeReplies toggles matching replies. Defaults to true.
	IncludeReplies *bool `json:"include_replies,omitempty"`
	// IncludeQuotes toggles matching quote tweets. Defaults to true.
	IncludeQuotes *bool `json:"include_quotes,omitempty"`
	// MinFaves is the minimum like count for a tweet to match. Defaults to 0.
	MinFaves int `json:"min_faves,omitempty"`
	// Keyword, if set, only matches tweets containing this substring.
	Keyword string `json:"keyword,omitempty"`
	// Lang, if set, only matches tweets in this language code.
	Lang string `json:"lang,omitempty"`
}

// XMonitor represents a single X (Twitter) monitor. The response from
// XMonitorCreate carries only the summary fields (id through active); the
// list and get endpoints populate the full object, including the match
// filters and the last_* / created_at timestamps.
type XMonitor struct {
	ID              string       `json:"id"`
	Kind            XMonitorKind `json:"kind"`
	Target          string       `json:"target"`
	Name            string       `json:"name,omitempty"`
	IntervalMinutes int          `json:"interval_minutes"`
	WebhookURL      string       `json:"webhook_url,omitempty"`
	Active          bool         `json:"active"`

	// The following fields are only populated by XMonitorList and XMonitorGet.
	IncludeRetweets bool   `json:"include_retweets,omitempty"`
	IncludeReplies  bool   `json:"include_replies,omitempty"`
	IncludeQuotes   bool   `json:"include_quotes,omitempty"`
	MinFaves        int    `json:"min_faves,omitempty"`
	Keyword         string `json:"keyword,omitempty"`
	Lang            string `json:"lang,omitempty"`
	LastCheckedAt   string `json:"last_checked_at,omitempty"`
	LastMatchedAt   string `json:"last_matched_at,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

// XMonitorListResponse is the paginated list of X monitors.
type XMonitorListResponse struct {
	Monitors []XMonitor `json:"monitors"`
}

// XMonitorUpdateRequest updates a mutable subset of a monitor's fields. All
// fields are optional; nil pointers are omitted so unset fields are left
// unchanged. Active is a pointer so a monitor can be explicitly paused
// (false) or resumed (true) rather than only ever enabled.
type XMonitorUpdateRequest struct {
	Name            *string `json:"name,omitempty"`
	IntervalMinutes *int    `json:"interval_minutes,omitempty"`
	WebhookURL      *string `json:"webhook_url,omitempty"`
	Active          *bool   `json:"active,omitempty"`
}

// XMonitorMutationResponse is the {"success": true} acknowledgement returned
// by XMonitorUpdate and XMonitorDelete.
type XMonitorMutationResponse struct {
	Success bool `json:"success"`
}

// XMonitorCheckResponse is returned when an immediate check is triggered. The
// check runs in the background, so Status is "checking" rather than a result.
type XMonitorCheckResponse struct {
	Status string `json:"status"`
}

// CreateXMonitor creates a new X (Twitter) monitor. It is the X analog of
// WatchCreate. Paid-only: a 403 is returned for free or lapsed accounts, and
// a user may hold at most 50 monitors.
func (c *Client) CreateXMonitor(ctx context.Context, req *XMonitorCreateRequest) (*XMonitor, error) {
	var resp XMonitor
	if err := c.do(ctx, "POST", "/v1/x/monitors", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListXMonitors returns the authenticated user's X monitors with pagination.
// limit is clamped server-side to 1..100 and offset must be >= 0.
func (c *Client) ListXMonitors(ctx context.Context, limit, offset int) (*XMonitorListResponse, error) {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	path := "/v1/x/monitors?" + q.Encode()

	var resp XMonitorListResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetXMonitor retrieves a single X monitor by ID, including its match filters
// and last-checked / last-matched timestamps.
func (c *Client) GetXMonitor(ctx context.Context, id string) (*XMonitor, error) {
	var resp XMonitor
	if err := c.do(ctx, "GET", fmt.Sprintf("/v1/x/monitors/%s", pathSegment(id)), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateXMonitor updates a monitor's name, interval, webhook URL, or active
// state. Only the non-nil fields of req are sent.
func (c *Client) UpdateXMonitor(ctx context.Context, id string, req *XMonitorUpdateRequest) (*XMonitorMutationResponse, error) {
	var resp XMonitorMutationResponse
	if err := c.do(ctx, "PATCH", fmt.Sprintf("/v1/x/monitors/%s", pathSegment(id)), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteXMonitor removes an X monitor.
func (c *Client) DeleteXMonitor(ctx context.Context, id string) (*XMonitorMutationResponse, error) {
	var resp XMonitorMutationResponse
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/v1/x/monitors/%s", pathSegment(id)), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CheckXMonitor triggers an immediate check of a monitor. The check runs in
// the background and costs one credit; the returned Status is "checking".
func (c *Client) CheckXMonitor(ctx context.Context, id string) (*XMonitorCheckResponse, error) {
	var resp XMonitorCheckResponse
	if err := c.do(ctx, "POST", fmt.Sprintf("/v1/x/monitors/%s/check", pathSegment(id)), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
