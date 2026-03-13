# webclaw-go

Go SDK for the [Webclaw](https://webclaw.io) web extraction API.

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

    webclaw "github.com/0xMassi/webclaw-go"
)

func main() {
    client := webclaw.NewClient("wc_your_api_key")

    result, err := client.Scrape(context.Background(), webclaw.ScrapeRequest{
        URL:     "https://example.com",
        Formats: []webclaw.Format{webclaw.FormatMarkdown},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Markdown)
}
```

## API Reference

### Scrape

Extract content from a single URL.

```go
result, err := client.Scrape(ctx, webclaw.ScrapeRequest{
    URL:              "https://example.com",
    Formats:          []webclaw.Format{webclaw.FormatMarkdown, webclaw.FormatText},
    IncludeSelectors: []string{"article", ".content"},
    ExcludeSelectors: []string{"nav", "footer"},
    OnlyMainContent:  true,
    NoCache:          true,
})
```

### Crawl

Start an async crawl job and poll for results.

```go
job, err := client.Crawl(ctx, webclaw.CrawlRequest{
    URL:        "https://example.com",
    MaxDepth:   3,
    MaxPages:   100,
    UseSitemap: true,
})
if err != nil {
    panic(err)
}

// Poll until complete (2s interval, 5min timeout)
status, err := client.WaitForCrawl(ctx, job.ID, 2*time.Second, 5*time.Minute)
if err != nil {
    panic(err)
}

for _, page := range status.Pages {
    fmt.Println(page.URL, len(page.Markdown))
}
```

### Map

Discover URLs via sitemap parsing.

```go
result, err := client.Map(ctx, webclaw.MapRequest{
    URL: "https://example.com",
})
fmt.Println(result.Count)
for _, u := range result.URLs {
    fmt.Println(u)
}
```

### Batch

Scrape multiple URLs in parallel.

```go
result, err := client.Batch(ctx, webclaw.BatchRequest{
    URLs:        []string{"https://a.com", "https://b.com"},
    Formats:     []webclaw.Format{webclaw.FormatMarkdown},
    Concurrency: 5,
})
for _, item := range result.Results {
    fmt.Println(item.URL, item.Error)
}
```

### Extract

LLM-powered structured data extraction.

```go
// Schema-based
result, err := client.Extract(ctx, webclaw.ExtractRequest{
    URL:    "https://example.com/pricing",
    Schema: json.RawMessage(`{"type":"object","properties":{"plans":{"type":"array"}}}`),
})
fmt.Println(string(result.Data))

// Prompt-based
result, err = client.Extract(ctx, webclaw.ExtractRequest{
    URL:    "https://example.com",
    Prompt: "Extract all pricing tiers",
})
```

### Summarize

```go
result, err := client.Summarize(ctx, webclaw.SummarizeRequest{
    URL:          "https://example.com",
    MaxSentences: 3,
})
fmt.Println(result.Summary)
```

### Brand

Extract brand identity (colors, fonts, logos).

```go
result, err := client.Brand(ctx, webclaw.BrandRequest{
    URL: "https://example.com",
})
fmt.Println(string(result.Data))
```

## Error Handling

```go
result, err := client.Scrape(ctx, req)
if err != nil {
    var apiErr *webclaw.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Configuration

```go
client := webclaw.NewClient(
    "wc_your_api_key",
    webclaw.WithBaseURL("https://api.webclaw.io"),
    webclaw.WithTimeout(60 * time.Second),
    webclaw.WithHTTPClient(customClient),
)
```

## Design

- Zero dependencies beyond the standard library
- `context.Context` on every method
- Functional options pattern for client configuration
- All API errors returned as `*APIError` with status code

## Requirements

- Go 1.21+

## License

MIT
