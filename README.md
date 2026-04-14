<p align="center">
  <a href="https://webclaw.io">
    <img src=".github/banner.png" alt="webclaw" width="600" />
  </a>
</p>

<p align="center">
  <strong>Go SDK for the Webclaw web extraction API</strong>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/0xMassi/webclaw-go"><img src="https://img.shields.io/badge/go-reference-212529?style=flat-square" alt="Go Reference" /></a>
  <a href="https://github.com/0xMassi/webclaw-go/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-212529?style=flat-square" alt="License" /></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/go-%3E%3D1.21-212529?style=flat-square" alt="Go 1.21+" /></a>
</p>

---

> **Note**: The webclaw Cloud API is currently in closed beta. [Request early access](https://webclaw.io) or use the [open-source CLI/MCP](https://github.com/0xMassi/webclaw) for local extraction.

---

## Installation

```bash
go get github.com/0xMassi/webclaw-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    webclaw "github.com/0xMassi/webclaw-go"
)

func main() {
    client := webclaw.NewClient(os.Getenv("WEBCLAW_API_KEY"))

    result, err := client.Scrape(context.Background(), &webclaw.ScrapeRequest{
        URL:     "https://example.com",
        Formats: []webclaw.Format{webclaw.FormatMarkdown},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Markdown)
}
```

## Highlights

- Zero dependencies beyond `net/http`
- `context.Context` on every method for cancellation and timeouts
- Functional options pattern for client configuration
- Typed errors with helper functions (`IsRateLimited`, `IsAuthError`, `IsNotFound`)
- Async polling helpers for crawl and research jobs

## Configuration

```go
client := webclaw.NewClient(
    os.Getenv("WEBCLAW_API_KEY"),
    webclaw.WithBaseURL("https://api.webclaw.io"),
    webclaw.WithTimeout(60 * time.Second),
    webclaw.WithHTTPClient(customHTTPClient),
)
```

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Override the default API base URL (`https://api.webclaw.io`) |
| `WithTimeout(d)` | Set the HTTP client timeout (default: 30s) |
| `WithHTTPClient(c)` | Replace the default `*http.Client` entirely |

## Endpoints

### Scrape

Extract content from a single URL.

```go
result, err := client.Scrape(ctx, &webclaw.ScrapeRequest{
    URL:              "https://example.com",
    Formats:          []webclaw.Format{webclaw.FormatMarkdown, webclaw.FormatText},
    IncludeSelectors: []string{"article", ".content"},
    ExcludeSelectors: []string{"nav", "footer"},
    OnlyMainContent:  true,
    NoCache:          true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Markdown)
fmt.Println(result.Cache.Status) // "hit", "miss", or "bypass"
```

### Search

Web search with optional scraping of results.

```go
resp, err := client.Search(ctx, &webclaw.SearchRequest{
    Query:      "web scraping tools 2026",
    NumResults: 10,
    Country:    "us",
})
if err != nil {
    log.Fatal(err)
}
for _, r := range resp.Results {
    fmt.Printf("%d. %s — %s\n", r.Position, r.Title, r.URL)
}
```

### Map

Discover URLs on a site via its sitemap.

```go
result, err := client.Map(ctx, &webclaw.MapRequest{
    URL: "https://example.com",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d URLs\n", result.Count)
for _, u := range result.URLs {
    fmt.Println(u)
}
```

### Batch

Scrape multiple URLs in parallel.

```go
result, err := client.Batch(ctx, &webclaw.BatchRequest{
    URLs:        []string{"https://a.com", "https://b.com", "https://c.com"},
    Formats:     []webclaw.Format{webclaw.FormatMarkdown},
    Concurrency: 5,
})
if err != nil {
    log.Fatal(err)
}
for _, item := range result.Results {
    if item.Error != "" {
        fmt.Printf("FAIL %s: %s\n", item.URL, item.Error)
        continue
    }
    fmt.Printf("OK   %s (%d bytes)\n", item.URL, len(item.Markdown))
}
```

### Extract

LLM-powered structured data extraction. Provide either a JSON schema or a natural-language prompt.

```go
// Schema-based extraction
result, err := client.Extract(ctx, &webclaw.ExtractRequest{
    URL:    "https://example.com/pricing",
    Schema: json.RawMessage(`{"type":"object","properties":{"plans":{"type":"array"}}}`),
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result.Data))

// Prompt-based extraction
result, err = client.Extract(ctx, &webclaw.ExtractRequest{
    URL:    "https://example.com/pricing",
    Prompt: "Extract all pricing tiers with name, price, and features",
})
```

### Summarize

Generate a plain-text summary of a page.

```go
result, err := client.Summarize(ctx, &webclaw.SummarizeRequest{
    URL:          "https://example.com/blog/long-article",
    MaxSentences: 3,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Summary)
```

### Brand

Extract brand identity information (colors, fonts, logos) from a URL. The response is a flexible JSON object since the shape depends on the target site.

```go
result, err := client.Brand(ctx, &webclaw.BrandRequest{
    URL: "https://example.com",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result.Data))

// Decode into a custom struct
var brand struct {
    Name   string   `json:"name"`
    Colors []string `json:"colors"`
}
if err := result.Decode(&brand); err != nil {
    log.Fatal(err)
}
fmt.Println(brand.Name, brand.Colors)
```

### Diff

Compare the current state of a page against a previous snapshot to detect changes.

```go
result, err := client.Diff(ctx, &webclaw.DiffRequest{
    URL: "https://example.com/pricing",
    Previous: map[string]interface{}{
        "title": "Old Pricing Page",
        "price": "$9.99",
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Changes)
```

### Agent Scrape

AI-guided scraping. Provide a goal and the agent navigates, clicks, and extracts data across multiple steps.

```go
result, err := client.AgentScrape(ctx, &webclaw.AgentScrapeRequest{
    URL:      "https://example.com/products",
    Goal:     "Find the price and specs of the top 3 products",
    MaxSteps: 10,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Completed in %d steps\n", result.TotalSteps)
fmt.Println(result.Data)
for _, step := range result.Steps {
    fmt.Printf("Step %d: %v\n", step.Step, step.Action)
}
```

### Research

Start an async deep research job and poll for results. Research can take several minutes depending on the query and configuration.

```go
// Start the job
job, err := client.Research(ctx, &webclaw.ResearchRequest{
    Query:      "How do modern web crawlers handle JavaScript rendering?",
    MaxSources: 15,
    Deep:       true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Research job started: %s\n", job.ID)

// Poll until complete (default: 2s interval, 10min timeout)
result, err := client.WaitForResearch(ctx, job.ID, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Report)
fmt.Printf("Sources: %d, Findings: %d\n", result.SourcesCount, result.FindingsCount)

// Or poll manually
status, err := client.GetResearchStatus(ctx, job.ID)

// Custom poll options
result, err = client.WaitForResearch(ctx, job.ID, &webclaw.ResearchPollOptions{
    Interval: 5 * time.Second,
    Timeout:  15 * time.Minute,
})
```

### Crawl

Start an async crawl job and poll until completion.

```go
// Start the crawl
job, err := client.Crawl(ctx, &webclaw.CrawlRequest{
    URL:        "https://example.com",
    MaxDepth:   3,
    MaxPages:   100,
    UseSitemap: true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Crawl started: %s\n", job.ID)

// Poll until complete (default: 2s interval, no timeout beyond parent context)
status, err := client.WaitForCompletion(ctx, job.ID, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Crawled %d/%d pages (%d errors)\n", status.Completed, status.Total, status.Errors)
for _, page := range status.Pages {
    if page.Error != "" {
        fmt.Printf("FAIL %s: %s\n", page.URL, page.Error)
        continue
    }
    fmt.Printf("OK   %s (%d bytes)\n", page.URL, len(page.Markdown))
}

// Or poll manually
status, err = client.GetCrawl(ctx, job.ID)

// Custom poll options
status, err = client.WaitForCompletion(ctx, job.ID, &webclaw.CrawlPollOptions{
    Interval: 5 * time.Second,
    Timeout:  10 * time.Minute,
})
```

### Watch

Monitor URLs for changes over time. Create watches, list them, check them manually, and clean up.

**Create a watch**

```go
watch, err := client.WatchCreate(ctx, &webclaw.WatchCreateRequest{
    URL:             "https://example.com/pricing",
    Name:            "Pricing page",
    IntervalMinutes: 60,
    WebhookURL:      "https://hooks.example.com/webclaw",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Watch created: %s (checks every %d min)\n", watch.ID, watch.IntervalMinutes)
```

**List watches**

```go
list, err := client.WatchList(ctx, 20, 0) // limit=20, offset=0
if err != nil {
    log.Fatal(err)
}
for _, w := range list.Watches {
    fmt.Printf("%s — %s (active: %v)\n", w.ID, w.URL, w.Active)
}
```

**Get watch details with snapshots**

```go
detail, err := client.WatchGet(ctx, "watch_id_here")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("URL: %s, Last changed: %s\n", detail.URL, detail.LastChangedAt)
for _, snap := range detail.Snapshots {
    fmt.Printf("  %s — %d words (delta: %+d)\n", snap.CheckedAt, snap.WordCount, snap.WordCountDelta)
}
```

**Trigger a manual check**

```go
resp, err := client.WatchCheck(ctx, "watch_id_here")
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Status)
```

**Delete a watch**

```go
err := client.WatchDelete(ctx, "watch_id_here")
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

All API errors are returned as `*webclaw.APIError` with the HTTP status code and message. Use the helper functions to check for common error types.

```go
result, err := client.Scrape(ctx, &webclaw.ScrapeRequest{URL: "https://example.com"})
if err != nil {
    if webclaw.IsRateLimited(err) {
        // Back off and retry
        log.Println("Rate limited, retrying after delay...")
    } else if webclaw.IsAuthError(err) {
        // Invalid or expired API key
        log.Fatal("Authentication failed. Check your WEBCLAW_API_KEY.")
    } else if webclaw.IsNotFound(err) {
        log.Println("Resource not found")
    } else {
        // Generic API error
        var apiErr *webclaw.APIError
        if errors.As(err, &apiErr) {
            log.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
        } else {
            // Network or other non-API error
            log.Printf("Request failed: %v\n", err)
        }
    }
    return
}
```

## All Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Scrape` | `(ctx, *ScrapeRequest) (*ScrapeResponse, error)` | Extract content from a URL |
| `Search` | `(ctx, *SearchRequest) (*SearchResponse, error)` | Web search |
| `Map` | `(ctx, *MapRequest) (*MapResponse, error)` | Discover URLs via sitemap |
| `Batch` | `(ctx, *BatchRequest) (*BatchResponse, error)` | Multi-URL parallel scrape |
| `Extract` | `(ctx, *ExtractRequest) (*ExtractResponse, error)` | LLM structured extraction |
| `Summarize` | `(ctx, *SummarizeRequest) (*SummarizeResponse, error)` | Page summarization |
| `Brand` | `(ctx, *BrandRequest) (*BrandResponse, error)` | Brand identity extraction |
| `Diff` | `(ctx, *DiffRequest) (*DiffResponse, error)` | Content change detection |
| `AgentScrape` | `(ctx, *AgentScrapeRequest) (*AgentScrapeResponse, error)` | AI-guided scraping |
| `Research` | `(ctx, *ResearchRequest) (*ResearchStartResponse, error)` | Start research job |
| `GetResearchStatus` | `(ctx, id) (*ResearchResponse, error)` | Poll research status |
| `WaitForResearch` | `(ctx, id, *ResearchPollOptions) (*ResearchResponse, error)` | Block until research completes |
| `Crawl` | `(ctx, *CrawlRequest) (*CrawlStartResponse, error)` | Start crawl job |
| `GetCrawl` | `(ctx, id) (*CrawlStatusResponse, error)` | Poll crawl status |
| `WaitForCompletion` | `(ctx, id, *CrawlPollOptions) (*CrawlStatusResponse, error)` | Block until crawl completes |
| `WatchCreate` | `(ctx, *WatchCreateRequest) (*WatchEntry, error)` | Create URL watch |
| `WatchList` | `(ctx, limit, offset) (*WatchListResponse, error)` | List watches |
| `WatchGet` | `(ctx, id) (*WatchDetail, error)` | Get watch with snapshots |
| `WatchDelete` | `(ctx, id) error` | Delete watch |
| `WatchCheck` | `(ctx, id) (*WatchCheckResponse, error)` | Trigger manual check |

## License

MIT
