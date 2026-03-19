// Package webclaw provides a Go SDK for the webclaw web extraction API.
package webclaw

import "encoding/json"

// Format represents a content format returned by the scrape endpoint.
type Format string

const (
	// FormatMarkdown requests Markdown output.
	FormatMarkdown Format = "markdown"
	// FormatText requests plain-text output.
	FormatText Format = "text"
	// FormatLLM requests LLM-optimized output (compressed markdown).
	FormatLLM Format = "llm"
	// FormatJSON requests structured JSON output.
	FormatJSON Format = "json"
)

// CrawlStatus represents the state of an async crawl job.
type CrawlStatus string

const (
	// CrawlStatusRunning indicates the crawl is still in progress.
	CrawlStatusRunning CrawlStatus = "running"
	// CrawlStatusCompleted indicates the crawl finished successfully.
	CrawlStatusCompleted CrawlStatus = "completed"
	// CrawlStatusFailed indicates the crawl encountered an unrecoverable error.
	CrawlStatusFailed CrawlStatus = "failed"
)

// CacheStatus represents whether a scrape result was served from cache.
type CacheStatus string

const (
	// CacheHit means the response was served from cache.
	CacheHit CacheStatus = "hit"
	// CacheMiss means no cached entry existed and a fresh fetch was performed.
	CacheMiss CacheStatus = "miss"
	// CacheBypass means caching was explicitly skipped via NoCache.
	CacheBypass CacheStatus = "bypass"
)

// --- Scrape ---

// ScrapeRequest configures a single URL scrape.
type ScrapeRequest struct {
	URL              string   `json:"url"`
	Formats          []Format `json:"formats,omitempty"`
	IncludeSelectors []string `json:"include_selectors,omitempty"`
	ExcludeSelectors []string `json:"exclude_selectors,omitempty"`
	OnlyMainContent  bool     `json:"only_main_content,omitempty"`
	NoCache          bool     `json:"no_cache,omitempty"`
}

// CacheInfo describes the cache status of a scrape response.
type CacheInfo struct {
	Status CacheStatus `json:"status"`
}

// ScrapeResponse contains the extracted content from a scrape.
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

// CrawlRequest configures an async crawl job.
type CrawlRequest struct {
	URL        string `json:"url"`
	MaxDepth   int    `json:"max_depth,omitempty"`
	MaxPages   int    `json:"max_pages,omitempty"`
	UseSitemap bool   `json:"use_sitemap,omitempty"`
}

// CrawlStartResponse is returned when a crawl job is created.
type CrawlStartResponse struct {
	ID     string      `json:"id"`
	Status CrawlStatus `json:"status"`
}

// CrawlPage holds the extracted content for one page in a crawl.
type CrawlPage struct {
	URL      string          `json:"url"`
	Markdown string          `json:"markdown,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// CrawlStatusResponse contains the current state and results of a crawl job.
type CrawlStatusResponse struct {
	ID        string      `json:"id"`
	Status    CrawlStatus `json:"status"`
	Pages     []CrawlPage `json:"pages,omitempty"`
	Total     int         `json:"total"`
	Completed int         `json:"completed"`
	Errors    int         `json:"errors"`
}

// --- Map ---

// MapRequest configures a sitemap URL discovery request.
type MapRequest struct {
	URL string `json:"url"`
}

// MapResponse contains the discovered URLs from a sitemap.
type MapResponse struct {
	URLs  []string `json:"urls"`
	Count int      `json:"count"`
}

// --- Batch ---

// BatchRequest configures a multi-URL parallel scrape.
type BatchRequest struct {
	URLs        []string `json:"urls"`
	Formats     []Format `json:"formats,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
}

// BatchResult holds the extracted content for one URL in a batch.
type BatchResult struct {
	URL      string          `json:"url"`
	Markdown string          `json:"markdown,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// BatchResponse contains the results of a batch scrape.
type BatchResponse struct {
	Results []BatchResult `json:"results"`
}

// --- Extract ---

// ExtractRequest configures an LLM-powered data extraction.
type ExtractRequest struct {
	URL    string          `json:"url"`
	Schema json.RawMessage `json:"schema,omitempty"`
	Prompt string          `json:"prompt,omitempty"`
}

// ExtractResponse contains the structured data returned by extraction.
type ExtractResponse struct {
	Data json.RawMessage `json:"data"`
}

// --- Summarize ---

// SummarizeRequest configures a page summarization request.
type SummarizeRequest struct {
	URL          string `json:"url"`
	MaxSentences int    `json:"max_sentences,omitempty"`
}

// SummarizeResponse contains the generated summary.
type SummarizeResponse struct {
	Summary string `json:"summary"`
}

// --- Brand ---

// BrandRequest configures a brand identity extraction request.
type BrandRequest struct {
	URL string `json:"url"`
}

// BrandResponse holds brand identity data as a flexible JSON object,
// since the shape depends on what the target site exposes.
type BrandResponse struct {
	Data json.RawMessage
}
