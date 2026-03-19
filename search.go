package webclaw

import "context"

// SearchRequest configures a web search query.
type SearchRequest struct {
	Query      string   `json:"query"`
	NumResults int      `json:"num_results,omitempty"`
	Scrape     *bool    `json:"scrape,omitempty"`
	Formats    []Format `json:"formats,omitempty"`
	Country    string   `json:"country,omitempty"`
	Lang       string   `json:"lang,omitempty"`
}

// SearchResult is a single search hit.
type SearchResult struct {
	Title    string         `json:"title"`
	URL      string         `json:"url"`
	Snippet  string         `json:"snippet"`
	Position int            `json:"position"`
	Markdown string         `json:"markdown,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SearchResponse is the result of a search request.
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Scrape  bool           `json:"scrape"`
}

// Search performs a web search query.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse
	if err := c.do(ctx, "POST", "/v1/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
