package webclaw

import (
	"context"
	"encoding/json"
	"net/url"
)

// ExtractorInfo is one catalog entry returned by ListExtractors.
// It describes a vertical extractor: its stable name (usable with
// ScrapeVertical), a human-friendly label, a one-line description,
// and the URL patterns it claims.
type ExtractorInfo struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	URLPatterns []string `json:"url_patterns"`
}

// ListExtractorsResponse is the shape of GET /v1/extractors.
type ListExtractorsResponse struct {
	Extractors []ExtractorInfo `json:"extractors"`
}

// VerticalScrapeResponse is the shape of POST /v1/scrape/{vertical}.
// The Data field is extractor-specific: its keys depend on which
// vertical ran. We keep it as json.RawMessage so callers can
// unmarshal into their own per-vertical struct, or walk it as a
// generic map[string]any via json.Unmarshal(&Data, &m).
type VerticalScrapeResponse struct {
	Vertical string          `json:"vertical"`
	URL      string          `json:"url"`
	Data     json.RawMessage `json:"data"`
}

// ListExtractors returns the full catalog of vertical extractors
// available on the server. Useful for building UIs or CLIs that
// let users pick an extractor by name.
//
// Each entry's Name is the identifier to pass to ScrapeVertical.
// The catalog is stable across releases; new verticals append at the
// end.
func (c *Client) ListExtractors(ctx context.Context) (*ListExtractorsResponse, error) {
	var resp ListExtractorsResponse
	if err := c.do(ctx, "GET", "/v1/extractors", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ScrapeVertical runs a site-specific extractor by name and returns
// typed JSON with fields specific to the target site (title, price,
// author, rating, etc.) rather than generic markdown.
//
// See ListExtractors for the full list of available names. Common
// examples: "reddit", "github_repo", "trustpilot_reviews",
// "youtube_video", "shopify_product".
//
// The response Data is a json.RawMessage because its shape varies per
// vertical. Callers that know which vertical they invoked can
// unmarshal into their own struct:
//
//	resp, _ := client.ScrapeVertical(ctx, "github_repo", "https://github.com/rust-lang/rust")
//	var repo struct { Name string; Stars int64; Description string }
//	json.Unmarshal(resp.Data, &repo)
func (c *Client) ScrapeVertical(ctx context.Context, name, targetURL string) (*VerticalScrapeResponse, error) {
	// URL-encode the name to be safe against caller typos even though
	// real names are always [a-z_].
	path := "/v1/scrape/" + url.PathEscape(name)
	body := map[string]string{"url": targetURL}
	var resp VerticalScrapeResponse
	if err := c.do(ctx, "POST", path, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
