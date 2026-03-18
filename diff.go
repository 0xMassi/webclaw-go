package webclaw

import "context"

// DiffRequest configures a page diff request.
type DiffRequest struct {
	URL      string                 `json:"url"`
	Previous map[string]interface{} `json:"previous,omitempty"`
}

// DiffResponse contains the detected changes.
type DiffResponse struct {
	URL     string                 `json:"url"`
	Changes map[string]interface{} `json:"changes"`
}

// Diff compares the current state of a page against a previous snapshot.
func (c *Client) Diff(ctx context.Context, req *DiffRequest) (*DiffResponse, error) {
	var resp DiffResponse
	if err := c.do(ctx, "POST", "/v1/diff", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
