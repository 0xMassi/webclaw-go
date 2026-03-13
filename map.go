package webclaw

import "context"

// Map discovers all URLs on a site.
func (c *Client) Map(ctx context.Context, req *MapRequest) (*MapResponse, error) {
	var resp MapResponse
	if err := c.do(ctx, "POST", "/v1/map", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
