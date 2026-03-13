// Package webclaw provides a Go SDK for the webclaw web extraction API.
package webclaw

import "encoding/json"

// Format represents a content format returned by the scrape endpoint.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatText     Format = "text"
	FormatLLM      Format = "llm"
	FormatJSON     Format = "json"
)

// CrawlStatus represents the state of an async crawl job.
type CrawlStatus string

const (
	CrawlStatusRunning   CrawlStatus = "running"
	CrawlStatusCompleted CrawlStatus = "completed"
	CrawlStatusFailed    CrawlStatus = "failed"
)

// CacheStatus represents whether a scrape result was cached.
type CacheStatus string

const (
	CacheHit    CacheStatus = "hit"
	CacheMiss   CacheStatus = "miss"
	CacheBypass CacheStatus = "bypass"
)

// --- Scrape ---

type ScrapeRequest struct {
	URL              string   `json:"url"`
	Formats          []Format `json:"formats,omitempty"`
	IncludeSelectors []string `json:"include_selectors,omitempty"`
	ExcludeSelectors []string `json:"exclude_selectors,omitempty"`
	OnlyMainContent  bool     `json:"only_main_content,omitempty"`
	NoCache          bool     `json:"no_cache,omitempty"`
}

type CacheInfo struct {
	Status CacheStatus `json:"status"`
}

type ScrapeResponse struct {
	URL      string          `json:"url"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Markdown string          `json:"markdown,omitempty"`
	Text     string          `json:"text,omitempty"`
	LLM      string          `json:"llm,omitempty"`
	Cache    CacheInfo       `json:"cache"`
	Warning  string          `json:"warning,omitempty"`
}

// --- Crawl ---

type CrawlRequest struct {
	URL        string `json:"url"`
	MaxDepth   int    `json:"max_depth,omitempty"`
	MaxPages   int    `json:"max_pages,omitempty"`
	UseSitemap bool   `json:"use_sitemap,omitempty"`
}

type CrawlStartResponse struct {
	ID     string      `json:"id"`
	Status CrawlStatus `json:"status"`
}

type CrawlPage struct {
	URL      string          `json:"url"`
	Markdown string          `json:"markdown,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type CrawlStatusResponse struct {
	ID        string      `json:"id"`
	Status    CrawlStatus `json:"status"`
	Pages     []CrawlPage `json:"pages,omitempty"`
	Total     int         `json:"total"`
	Completed int         `json:"completed"`
	Errors    int         `json:"errors"`
}

// --- Map ---

type MapRequest struct {
	URL string `json:"url"`
}

type MapResponse struct {
	URLs  []string `json:"urls"`
	Count int      `json:"count"`
}

// --- Batch ---

type BatchRequest struct {
	URLs        []string `json:"urls"`
	Formats     []Format `json:"formats,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
}

type BatchResult struct {
	URL      string          `json:"url"`
	Markdown string          `json:"markdown,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Error    string          `json:"error,omitempty"`
}

type BatchResponse struct {
	Results []BatchResult `json:"results"`
}

// --- Extract ---

type ExtractRequest struct {
	URL    string          `json:"url"`
	Schema json.RawMessage `json:"schema,omitempty"`
	Prompt string          `json:"prompt,omitempty"`
}

type ExtractResponse struct {
	Data json.RawMessage `json:"data"`
}

// --- Summarize ---

type SummarizeRequest struct {
	URL          string `json:"url"`
	MaxSentences int    `json:"max_sentences,omitempty"`
}

type SummarizeResponse struct {
	Summary string `json:"summary"`
}

// --- Brand ---

type BrandRequest struct {
	URL string `json:"url"`
}

// BrandResponse holds brand identity data as a flexible JSON object,
// since the shape depends on what the target site exposes.
type BrandResponse struct {
	Data json.RawMessage
}
