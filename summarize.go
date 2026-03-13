package webclaw

import "context"

// Summarize generates a plain-text summary of a URL's content.
func (c *Client) Summarize(ctx context.Context, req *SummarizeRequest) (*SummarizeResponse, error) {
	var resp SummarizeResponse
	if err := c.do(ctx, "POST", "/v1/summarize", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
