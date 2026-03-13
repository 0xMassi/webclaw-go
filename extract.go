package webclaw

import "context"

// Extract pulls structured data from a URL, optionally guided by a
// JSON schema or a natural-language prompt.
func (c *Client) Extract(ctx context.Context, req *ExtractRequest) (*ExtractResponse, error) {
	var resp ExtractResponse
	if err := c.do(ctx, "POST", "/v1/extract", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
