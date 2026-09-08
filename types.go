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
	FormatJSON       Format = "json"
	FormatExtract    Format = "extract"
	FormatLinks      Format = "links"
	FormatRawHTML    Format = "rawHtml"
	FormatAttributes Format = "attributes"
	FormatQuery      Format = "query"
)

// CrawlStatus represents the state of an async crawl job.
type CrawlStatus string

const (
	// CrawlStatusRunning indicates the crawl is still in progress.
	CrawlStatusRunning     CrawlStatus = "running"
	CrawlStatusPending     CrawlStatus = "pending"
	CrawlStatusInterrupted CrawlStatus = "interrupted"
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
	CacheSkip   CacheStatus = "skip"
)

// --- Scrape ---

// ScrapeRequest configures a single URL scrape.
type ExtractOptions struct {
	Schema map[string]any `json:"schema,omitempty"`
	Prompt string         `json:"prompt,omitempty"`
}

type ScrapeRequest struct {
	Extract            *ExtractOptions     `json:"extract,omitempty"`
	URL                string              `json:"url"`
	Formats            []Format            `json:"formats,omitempty"`
	IncludeSelectors   []string            `json:"include_selectors,omitempty"`
	ExcludeSelectors   []string            `json:"exclude_selectors,omitempty"`
	OnlyMainContent    bool                `json:"only_main_content,omitempty"`
	NoCache            bool                `json:"no_cache,omitempty"`
	MaxCacheAge        *uint64             `json:"max_cache_age,omitempty"`
	Mobile             bool                `json:"mobile,omitempty"`
	Screenshot         bool                `json:"screenshot,omitempty"`
	Actions            []map[string]any    `json:"actions,omitempty"`
	Query              string              `json:"query,omitempty"`
	AttributeSelectors []AttributeSelector `json:"attribute_selectors,omitempty"`
}

type AttributeSelector struct {
	Selector  string `json:"selector"`
	Attribute string `json:"attribute"`
}

// CacheInfo describes the cache status of a scrape response.
type CacheInfo struct {
	Status     CacheStatus `json:"status"`
	CachedAt   *string     `json:"cached_at,omitempty"`
	AgeSeconds *uint64     `json:"age_seconds,omitempty"`
}

// YouTubeData holds structured metadata for a YouTube watch URL,
// returned by /v1/scrape. Metadata may be available without captions;
// inspect Warning and Transcript before treating a response as a transcript.
type YouTubeData struct {
	VideoID         string   `json:"video_id,omitempty"`
	Title           string   `json:"title,omitempty"`
	Description     string   `json:"description,omitempty"`
	Channel         string   `json:"channel,omitempty"`
	ChannelURL      string   `json:"channel_url,omitempty"`
	Uploader        string   `json:"uploader,omitempty"`
	UploadDate      string   `json:"upload_date,omitempty"` // YYYYMMDD
	DurationSeconds int      `json:"duration_seconds,omitempty"`
	ViewCount       int      `json:"view_count,omitempty"`
	LikeCount       int      `json:"like_count,omitempty"`
	Thumbnail       string   `json:"thumbnail,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Categories      []string `json:"categories,omitempty"`
	Language        string   `json:"language,omitempty"`
}

// ScrapeResponse contains the extracted content from a scrape.
type ScrapeResponse struct {
	Extract any `json:"extract,omitempty"`
	// Extraction contains the full result requested by FormatJSON.
	Extraction json.RawMessage `json:"extraction,omitempty"`
	URL        string          `json:"url"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Markdown   string          `json:"markdown,omitempty"`
	Text       string          `json:"text,omitempty"`
	LLM        string          `json:"llm,omitempty"`
	Cache      *CacheInfo      `json:"cache"`
	Warning    string          `json:"warning,omitempty"`
	// YouTube is set when the URL is youtube.com/watch, /shorts, or
	// youtu.be. Carries channel, duration, view count, tags, etc.
	YouTube *YouTubeData `json:"youtube,omitempty"`
	// Transcript carries the auto-caption text (newline-joined). Only
	// present when the yt-dlp short-circuit fired and the video has
	// captions.
	Transcript       string           `json:"transcript,omitempty"`
	Links            []map[string]any `json:"links,omitempty"`
	RawHTML          string           `json:"rawHtml,omitempty"`
	Attributes       []map[string]any `json:"attributes,omitempty"`
	QueryAnswer      *string          `json:"query_answer,omitempty"`
	Screenshot       string           `json:"screenshot,omitempty"`
	ActionsPerformed int              `json:"actions_performed,omitempty"`
	Mobile           bool             `json:"mobile,omitempty"`
	StructuredData   json.RawMessage  `json:"structured_data,omitempty"`
	Engine           json.RawMessage  `json:"engine,omitempty"`
}

// --- Crawl ---

// CrawlRequest configures an async crawl job.
type CrawlRequest struct {
	IncludePatterns    []string `json:"include_patterns,omitempty"`
	ExcludePatterns    []string `json:"exclude_patterns,omitempty"`
	WebhookURL         string   `json:"webhook_url,omitempty"`
	AllowSubdomains    bool     `json:"allow_subdomains,omitempty"`
	AllowExternalLinks bool     `json:"allow_external_links,omitempty"`
	URL                string   `json:"url"`
	MaxDepth           int      `json:"max_depth,omitempty"`
	MaxPages           int      `json:"max_pages,omitempty"`
	UseSitemap         bool     `json:"use_sitemap,omitempty"`
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
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	URL    string `json:"url"`
}

// MapResponse contains the discovered URLs from a sitemap.
type MapResponse struct {
	NextCursor   *string  `json:"next_cursor,omitempty"`
	TotalIndexed int      `json:"total_indexed,omitempty"`
	Cached       bool     `json:"cached"`
	URLs         []string `json:"urls"`
	Count        int      `json:"count"`
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
	Extraction json.RawMessage `json:"extraction,omitempty"`
	Text       string          `json:"text,omitempty"`
	LLM        string          `json:"llm,omitempty"`
	URL        string          `json:"url"`
	Markdown   string          `json:"markdown,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Error      string          `json:"error,omitempty"`
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
