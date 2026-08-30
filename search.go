package webclaw

import "context"

// SearchFreshness is a provider discovery hint; result dates are not verified.
type SearchFreshness string

const (
	SearchFreshnessHour  SearchFreshness = "hour"
	SearchFreshnessDay   SearchFreshness = "day"
	SearchFreshnessWeek  SearchFreshness = "week"
	SearchFreshnessMonth SearchFreshness = "month"
	SearchFreshnessYear  SearchFreshness = "year"
)

// SearchRequest configures a web search query.
type SearchRequest struct {
	Query              string          `json:"query"`
	NumResults         int             `json:"num_results,omitempty"`
	Scrape             *bool           `json:"scrape,omitempty"`
	Formats            []Format        `json:"formats,omitempty"`
	Country            string          `json:"country,omitempty"`
	Lang               string          `json:"lang,omitempty"`
	IncludeDomains     []string        `json:"include_domains,omitempty"`
	ExcludeDomains     []string        `json:"exclude_domains,omitempty"`
	IncludeURLPrefixes []string        `json:"include_url_prefixes,omitempty"`
	Freshness          SearchFreshness `json:"freshness,omitempty"`
	// PublishedAfter is a provider discovery hint in YYYY-MM-DD format.
	PublishedAfter string `json:"published_after,omitempty"`
	// PublishedBefore is an exclusive provider discovery hint in YYYY-MM-DD format.
	PublishedBefore string `json:"published_before,omitempty"`
	Page            int    `json:"page,omitempty"`
	Location        string `json:"location,omitempty"`
	Autocorrect     *bool  `json:"autocorrect,omitempty"`
	NoCache         bool   `json:"no_cache,omitempty"`
	MaxCacheAge     int    `json:"max_cache_age,omitempty"`
}

// SearchAppliedFilters reports non-default source filters and provider hints.
type SearchAppliedFilters struct {
	IncludeDomains     []string        `json:"include_domains,omitempty"`
	ExcludeDomains     []string        `json:"exclude_domains,omitempty"`
	IncludeURLPrefixes []string        `json:"include_url_prefixes,omitempty"`
	Freshness          SearchFreshness `json:"freshness,omitempty"`
	PublishedAfter     string          `json:"published_after,omitempty"`
	PublishedBefore    string          `json:"published_before,omitempty"`
	Location           string          `json:"location,omitempty"`
	Autocorrect        *bool           `json:"autocorrect,omitempty"`
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
	Query            string                `json:"query"`
	Results          []SearchResult        `json:"results"`
	Scrape           bool                  `json:"scrape"`
	AppliedFilters   *SearchAppliedFilters `json:"applied_filters,omitempty"`
	FilteredOutCount int                   `json:"filtered_out_count,omitempty"`
	Page             int                   `json:"page,omitempty"`
}

// Search performs a web search query.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse
	if err := c.do(ctx, "POST", "/v1/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
