package webclaw

import "context"

// Batch scrapes multiple URLs in a single request.
func (c *Client) Batch(ctx context.Context, req *BatchRequest) (*BatchResponse, error) {
	var resp BatchResponse
	if err := c.do(ctx, "POST", "/v1/batch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
