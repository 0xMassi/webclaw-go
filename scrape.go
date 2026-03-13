package webclaw

import "context"

// Scrape fetches and converts a single URL.
func (c *Client) Scrape(ctx context.Context, req *ScrapeRequest) (*ScrapeResponse, error) {
	var resp ScrapeResponse
	if err := c.do(ctx, "POST", "/v1/scrape", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
