package webclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestServer returns an httptest.Server and a Client pointed at it.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := NewClient("test-key", WithBaseURL(srv.URL))
	return srv, client
}

// assertAuth checks that the request carries the expected Bearer token.
func assertAuth(t *testing.T, r *http.Request) {
	t.Helper()
	want := "Bearer test-key"
	if got := r.Header.Get("Authorization"); got != want {
		t.Errorf("auth header = %q, want %q", got, want)
	}
}

// --- Client construction ---

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("key123")
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.http.Timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", c.http.Timeout, defaultTimeout)
	}
	if c.apiKey != "key123" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "key123")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	c := NewClient("key", WithBaseURL("https://custom.io/"), WithTimeout(10*time.Second))
	if c.baseURL != "https://custom.io" {
		t.Errorf("baseURL = %q, want trailing slash stripped", c.baseURL)
	}
	if c.http.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", c.http.Timeout)
	}
}

// --- Scrape ---

func TestScrape_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/v1/scrape" || r.Method != "POST" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}

		var req ScrapeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.URL != "https://example.com" {
			t.Errorf("req.URL = %q", req.URL)
		}
		if len(req.Formats) != 2 {
			t.Errorf("req.Formats = %v", req.Formats)
		}

		json.NewEncoder(w).Encode(ScrapeResponse{
			URL:      "https://example.com",
			Markdown: "# Hello",
			Text:     "Hello",
			Cache:    CacheInfo{Status: CacheMiss},
		})
	})

	resp, err := client.Scrape(context.Background(), &ScrapeRequest{
		URL:     "https://example.com",
		Formats: []Format{FormatMarkdown, FormatText},
	})
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if resp.Markdown != "# Hello" {
		t.Errorf("Markdown = %q", resp.Markdown)
	}
	if resp.Cache.Status != CacheMiss {
		t.Errorf("Cache.Status = %q", resp.Cache.Status)
	}
}

func TestScrape_WithSelectors(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		var req ScrapeRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.IncludeSelectors) != 1 || req.IncludeSelectors[0] != "main" {
			t.Errorf("IncludeSelectors = %v", req.IncludeSelectors)
		}
		if len(req.ExcludeSelectors) != 1 || req.ExcludeSelectors[0] != "nav" {
			t.Errorf("ExcludeSelectors = %v", req.ExcludeSelectors)
		}
		if !req.OnlyMainContent {
			t.Error("OnlyMainContent should be true")
		}
		if !req.NoCache {
			t.Error("NoCache should be true")
		}

		json.NewEncoder(w).Encode(ScrapeResponse{URL: req.URL, Cache: CacheInfo{Status: CacheBypass}})
	})

	resp, err := client.Scrape(context.Background(), &ScrapeRequest{
		URL:              "https://example.com",
		IncludeSelectors: []string{"main"},
		ExcludeSelectors: []string{"nav"},
		OnlyMainContent:  true,
		NoCache:          true,
	})
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if resp.Cache.Status != CacheBypass {
		t.Errorf("Cache.Status = %q, want bypass", resp.Cache.Status)
	}
}

// --- Crawl ---

func TestCrawl_StartAndPoll(t *testing.T) {
	var pollCount atomic.Int32

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/crawl":
			json.NewEncoder(w).Encode(CrawlStartResponse{ID: "job-123", Status: CrawlStatusRunning})

		case r.Method == "GET" && r.URL.Path == "/v1/crawl/job-123":
			n := pollCount.Add(1)
			if n < 3 {
				json.NewEncoder(w).Encode(CrawlStatusResponse{
					ID: "job-123", Status: CrawlStatusRunning,
					Total: 5, Completed: int(n),
				})
			} else {
				json.NewEncoder(w).Encode(CrawlStatusResponse{
					ID: "job-123", Status: CrawlStatusCompleted,
					Total: 5, Completed: 5,
					Pages: []CrawlPage{
						{URL: "https://example.com", Markdown: "# Home"},
						{URL: "https://example.com/about", Markdown: "# About"},
					},
				})
			}

		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	start, err := client.Crawl(context.Background(), &CrawlRequest{
		URL:      "https://example.com",
		MaxDepth: 2,
		MaxPages: 50,
	})
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if start.ID != "job-123" {
		t.Errorf("ID = %q", start.ID)
	}

	result, err := client.WaitForCompletion(context.Background(), start.ID, &CrawlPollOptions{
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("WaitForCompletion: %v", err)
	}
	if result.Status != CrawlStatusCompleted {
		t.Errorf("Status = %q", result.Status)
	}
	if len(result.Pages) != 2 {
		t.Errorf("Pages count = %d", len(result.Pages))
	}
}

func TestWaitForCompletion_ContextCancelled(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Always return running -- we'll cancel the context.
		json.NewEncoder(w).Encode(CrawlStatusResponse{ID: "job-1", Status: CrawlStatusRunning})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.WaitForCompletion(ctx, "job-1", &CrawlPollOptions{
		Interval: 30 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected context error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}

func TestWaitForCompletion_NilOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CrawlStatusResponse{
			ID: "job-nil", Status: CrawlStatusCompleted,
			Total: 1, Completed: 1,
		})
	})

	result, err := client.WaitForCompletion(context.Background(), "job-nil", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != CrawlStatusCompleted {
		t.Errorf("Status = %q, want completed", result.Status)
	}
}

func TestWaitForCompletion_WithTimeout(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CrawlStatusResponse{ID: "job-slow", Status: CrawlStatusRunning})
	})

	_, err := client.WaitForCompletion(context.Background(), "job-slow", &CrawlPollOptions{
		Interval: 30 * time.Millisecond,
		Timeout:  100 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}

func TestWaitForCompletion_FailedJob(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CrawlStatusResponse{
			ID: "job-fail", Status: CrawlStatusFailed,
			Total: 3, Completed: 1, Errors: 2,
		})
	})

	result, err := client.WaitForCompletion(context.Background(), "job-fail", &CrawlPollOptions{
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != CrawlStatusFailed {
		t.Errorf("Status = %q, want failed", result.Status)
	}
	if result.Errors != 2 {
		t.Errorf("Errors = %d, want 2", result.Errors)
	}
}

func TestGetCrawl_NotFound(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "crawl not found"})
	})

	_, err := client.GetCrawl(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found, got: %v", err)
	}
}

// --- Map ---

func TestMap_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/v1/map" {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(MapResponse{
			URLs:  []string{"https://example.com", "https://example.com/about"},
			Count: 2,
		})
	})

	resp, err := client.Map(context.Background(), &MapRequest{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("Count = %d", resp.Count)
	}
	if len(resp.URLs) != 2 {
		t.Errorf("URLs = %v", resp.URLs)
	}
}

// --- Endpoints ---

func TestEndpoints_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/v1/endpoints" || r.Method != "POST" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}

		var req EndpointsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.URL != "https://example.com" {
			t.Errorf("req.URL = %q", req.URL)
		}

		json.NewEncoder(w).Encode(EndpointsResponse{
			URL:            "https://example.com",
			BundlesScanned: 4,
			EndpointCount:  3,
			Endpoints: []DiscoveredEndpoint{
				{Value: "/api/users", Kind: EndpointKindRelativePath, FirstParty: true, Source: "inline"},
				{Value: "https://api.example.com/v2", Kind: EndpointKindAbsoluteURL, FirstParty: true, Source: "app.js"},
				{Value: "wss://live.example.com/socket", Kind: EndpointKindWebSocket, FirstParty: false, Source: "app.js"},
			},
			Hosts:     []string{"api.example.com", "live.example.com"},
			Truncated: false,
		})
	})

	resp, err := client.Endpoints(context.Background(), &EndpointsRequest{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("Endpoints: %v", err)
	}
	if resp.BundlesScanned != 4 {
		t.Errorf("BundlesScanned = %d", resp.BundlesScanned)
	}
	if resp.EndpointCount != 3 || len(resp.Endpoints) != 3 {
		t.Errorf("EndpointCount = %d, len(Endpoints) = %d", resp.EndpointCount, len(resp.Endpoints))
	}
	if resp.Endpoints[0].Kind != EndpointKindRelativePath {
		t.Errorf("Endpoints[0].Kind = %q", resp.Endpoints[0].Kind)
	}
	if resp.Endpoints[2].Kind != EndpointKindWebSocket || resp.Endpoints[2].FirstParty {
		t.Errorf("Endpoints[2] = %+v", resp.Endpoints[2])
	}
	if len(resp.Hosts) != 2 {
		t.Errorf("Hosts = %v", resp.Hosts)
	}
}

func TestEndpoints_Params(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		var req EndpointsRequest
		json.NewDecoder(r.Body).Decode(&req)

		if !req.IncludeThirdParty {
			t.Error("IncludeThirdParty should be true")
		}
		if req.MaxBundles != 20 {
			t.Errorf("MaxBundles = %d, want 20", req.MaxBundles)
		}

		json.NewEncoder(w).Encode(EndpointsResponse{
			URL:           req.URL,
			EndpointCount: 0,
			Endpoints:     []DiscoveredEndpoint{},
			Truncated:     true,
		})
	})

	resp, err := client.Endpoints(context.Background(), &EndpointsRequest{
		URL:               "https://example.com",
		IncludeThirdParty: true,
		MaxBundles:        20,
	})
	if err != nil {
		t.Fatalf("Endpoints: %v", err)
	}
	if !resp.Truncated {
		t.Error("Truncated should be true")
	}
}

func TestEndpoints_BadRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "url is required"})
	})

	_, err := client.Endpoints(context.Background(), &EndpointsRequest{URL: ""})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Message != "url is required" {
		t.Errorf("Message = %q, want 'url is required'", apiErr.Message)
	}
}

// --- Batch ---

func TestBatch_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		var req BatchRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.URLs) != 2 {
			t.Errorf("URLs = %v", req.URLs)
		}
		if req.Concurrency != 3 {
			t.Errorf("Concurrency = %d", req.Concurrency)
		}

		json.NewEncoder(w).Encode(BatchResponse{
			Results: []BatchResult{
				{URL: req.URLs[0], Markdown: "# Page 1"},
				{URL: req.URLs[1], Error: "timeout"},
			},
		})
	})

	resp, err := client.Batch(context.Background(), &BatchRequest{
		URLs:        []string{"https://a.com", "https://b.com"},
		Formats:     []Format{FormatMarkdown},
		Concurrency: 3,
	})
	if err != nil {
		t.Fatalf("Batch: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("Results = %d", len(resp.Results))
	}
	if resp.Results[0].Markdown != "# Page 1" {
		t.Errorf("result[0].Markdown = %q", resp.Results[0].Markdown)
	}
	if resp.Results[1].Error != "timeout" {
		t.Errorf("result[1].Error = %q", resp.Results[1].Error)
	}
}

// --- Extract ---

func TestExtract_WithPrompt(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		var req ExtractRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Prompt != "get the title" {
			t.Errorf("Prompt = %q", req.Prompt)
		}

		json.NewEncoder(w).Encode(ExtractResponse{
			Data: json.RawMessage(`{"title":"Example"}`),
		})
	})

	resp, err := client.Extract(context.Background(), &ExtractRequest{
		URL:    "https://example.com",
		Prompt: "get the title",
	})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	var data struct{ Title string }
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Title != "Example" {
		t.Errorf("Title = %q", data.Title)
	}
}

func TestExtract_WithSchema(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"price":{"type":"number"}}}`)

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req ExtractRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Schema == nil {
			t.Error("Schema should not be nil")
		}

		json.NewEncoder(w).Encode(ExtractResponse{
			Data: json.RawMessage(`{"price":29.99}`),
		})
	})

	resp, err := client.Extract(context.Background(), &ExtractRequest{
		URL:    "https://shop.com",
		Schema: schema,
	})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	var data struct{ Price float64 }
	json.Unmarshal(resp.Data, &data)
	if data.Price != 29.99 {
		t.Errorf("Price = %f", data.Price)
	}
}

// --- Lead ---

func TestLead_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/v1/lead" {
			t.Errorf("path = %q, want /v1/lead", r.URL.Path)
		}

		var req LeadRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.URL != "https://posthog.com" {
			t.Errorf("URL = %q, want https://posthog.com", req.URL)
		}

		linkedin := "https://www.linkedin.com/in/james-hawkins"
		x := "https://x.com/james406"
		json.NewEncoder(w).Encode(LeadResponse{
			URL:    req.URL,
			Domain: "posthog.com",
			Lead: LeadData{
				CompanyName: "PostHog",
				Summary:     "Open-source product analytics",
				Socials:     LeadSocials{LinkedIn: "https://www.linkedin.com/company/posthog", X: "https://x.com/posthog", GitHub: "https://github.com/PostHog"},
				Tech:        []string{"React", "Django"},
				Pricing:     []LeadPricing{{Plan: "Free", Price: "$0"}},
				Emails:      []LeadEmail{{Type: "sales", Email: "hey@posthog.com"}},
				People:      []LeadPerson{{Name: "James Hawkins", Role: "CEO", LinkedIn: &linkedin, X: &x}},
			},
			PeopleSource: "team_page",
			Cache:        "miss",
			Credits:      100,
		})
	})

	resp, err := client.Lead(context.Background(), &LeadRequest{
		URL: "https://posthog.com",
	})
	if err != nil {
		t.Fatalf("Lead: %v", err)
	}
	if resp.Domain != "posthog.com" {
		t.Errorf("Domain = %q", resp.Domain)
	}
	if resp.Lead.CompanyName != "PostHog" {
		t.Errorf("CompanyName = %q", resp.Lead.CompanyName)
	}
	if resp.Lead.Socials.GitHub != "https://github.com/PostHog" {
		t.Errorf("Socials.GitHub = %q", resp.Lead.Socials.GitHub)
	}
	if len(resp.Lead.Emails) != 1 || resp.Lead.Emails[0].Email != "hey@posthog.com" {
		t.Errorf("Emails = %+v", resp.Lead.Emails)
	}
	if len(resp.Lead.People) != 1 || resp.Lead.People[0].Role != "CEO" {
		t.Fatalf("People = %+v", resp.Lead.People)
	}
	if resp.Lead.People[0].LinkedIn == nil || *resp.Lead.People[0].LinkedIn != "https://www.linkedin.com/in/james-hawkins" {
		t.Errorf("People[0].LinkedIn = %v", resp.Lead.People[0].LinkedIn)
	}
	if resp.Lead.People[0].X == nil || *resp.Lead.People[0].X != "https://x.com/james406" {
		t.Errorf("People[0].X = %v", resp.Lead.People[0].X)
	}
	if resp.PeopleSource != "team_page" {
		t.Errorf("PeopleSource = %q, want team_page", resp.PeopleSource)
	}
	if resp.Credits != 100 {
		t.Errorf("Credits = %d, want 100", resp.Credits)
	}
}

// --- Lead batch ---

func TestLeadBatch_StartAndPoll(t *testing.T) {
	var pollCount atomic.Int32

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/lead/batch":
			var req LeadBatchRequest
			json.NewDecoder(r.Body).Decode(&req)
			if len(req.URLs) != 2 {
				t.Errorf("URLs = %v", req.URLs)
			}
			if !req.NoCache {
				t.Error("NoCache should be true")
			}
			json.NewEncoder(w).Encode(LeadBatchStartResponse{
				ID: "lb-123", Status: "processing", Total: 2, CreditsPerURL: 100,
			})

		case r.Method == "GET" && r.URL.Path == "/v1/lead/batch/lb-123":
			n := pollCount.Add(1)
			if n < 3 {
				json.NewEncoder(w).Encode(LeadBatchResponse{
					ID: "lb-123", Status: "processing",
					Total: 2, Completed: int(n),
				})
			} else {
				linkedin := "https://www.linkedin.com/in/james-hawkins"
				json.NewEncoder(w).Encode(LeadBatchResponse{
					ID: "lb-123", Status: "completed",
					Total: 2, Completed: 2, Succeeded: 1, CreditsCharged: 100,
					Results: []LeadBatchResult{
						{
							URL: "https://posthog.com", Status: "success", Domain: "posthog.com", Cache: "miss",
							Lead: LeadData{
								CompanyName: "PostHog",
								Socials:     LeadSocials{GitHub: "https://github.com/PostHog"},
								People:      []LeadPerson{{Name: "James Hawkins", Role: "CEO", LinkedIn: &linkedin}},
							},
						},
						{URL: "https://broken.example", Status: "error", Error: "fetch failed"},
					},
					CreatedAt: "2026-07-18T12:00:00Z",
				})
			}

		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	start, err := client.LeadBatch(context.Background(), &LeadBatchRequest{
		URLs:    []string{"https://posthog.com", "https://broken.example"},
		NoCache: true,
	})
	if err != nil {
		t.Fatalf("LeadBatch: %v", err)
	}
	if start.ID != "lb-123" {
		t.Errorf("ID = %q", start.ID)
	}
	if start.Total != 2 {
		t.Errorf("Total = %d, want 2", start.Total)
	}
	if start.CreditsPerURL != 100 {
		t.Errorf("CreditsPerURL = %d, want 100", start.CreditsPerURL)
	}

	result, err := client.WaitForLeadBatch(context.Background(), start.ID, &LeadBatchPollOptions{
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("WaitForLeadBatch: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
	if result.Succeeded != 1 || result.CreditsCharged != 100 {
		t.Errorf("Succeeded = %d, CreditsCharged = %d", result.Succeeded, result.CreditsCharged)
	}
	if len(result.Results) != 2 {
		t.Fatalf("Results count = %d, want 2", len(result.Results))
	}

	ok := result.Results[0]
	if ok.Status != "success" || ok.Domain != "posthog.com" || ok.Cache != "miss" {
		t.Errorf("success result = %+v", ok)
	}
	if ok.Lead.CompanyName != "PostHog" {
		t.Errorf("Lead.CompanyName = %q", ok.Lead.CompanyName)
	}
	if len(ok.Lead.People) != 1 || ok.Lead.People[0].LinkedIn == nil {
		t.Errorf("Lead.People = %+v", ok.Lead.People)
	}

	bad := result.Results[1]
	if bad.Status != "error" || bad.Error != "fetch failed" {
		t.Errorf("error result = %+v", bad)
	}
}

func TestLeadBatch_InsufficientCredits(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(402)
		json.NewEncoder(w).Encode(map[string]string{"error": "insufficient credits"})
	})

	_, err := client.LeadBatch(context.Background(), &LeadBatchRequest{
		URLs: []string{"https://posthog.com"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 402 {
		t.Errorf("StatusCode = %d, want 402", apiErr.StatusCode)
	}
	if apiErr.Message != "insufficient credits" {
		t.Errorf("Message = %q, want 'insufficient credits'", apiErr.Message)
	}
}

func TestGetLeadBatch_NotFound(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "batch not found"})
	})

	_, err := client.GetLeadBatch(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found, got: %v", err)
	}
}

// --- Summarize ---

func TestSummarize_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		var req SummarizeRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.MaxSentences != 3 {
			t.Errorf("MaxSentences = %d", req.MaxSentences)
		}

		json.NewEncoder(w).Encode(SummarizeResponse{
			Summary: "This is a summary.",
		})
	})

	resp, err := client.Summarize(context.Background(), &SummarizeRequest{
		URL:          "https://example.com",
		MaxSentences: 3,
	})
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if resp.Summary != "This is a summary." {
		t.Errorf("Summary = %q", resp.Summary)
	}
}

// --- Brand ---

func TestBrand_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"Acme","logo":"https://acme.com/logo.png","colors":["#fff","#000"]}`)
	})

	resp, err := client.Brand(context.Background(), &BrandRequest{URL: "https://acme.com"})
	if err != nil {
		t.Fatalf("Brand: %v", err)
	}

	var brand struct {
		Name   string   `json:"name"`
		Logo   string   `json:"logo"`
		Colors []string `json:"colors"`
	}
	if err := resp.Decode(&brand); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if brand.Name != "Acme" {
		t.Errorf("Name = %q", brand.Name)
	}
	if len(brand.Colors) != 2 {
		t.Errorf("Colors = %v", brand.Colors)
	}
}

// --- Error handling ---

func TestAPIError_401(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid api key"})
	})

	_, err := client.Scrape(context.Background(), &ScrapeRequest{URL: "https://x.com"})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "invalid api key" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if !IsAuthError(err) {
		t.Error("IsAuthError should be true")
	}
	if IsRateLimited(err) {
		t.Error("IsRateLimited should be false")
	}
}

func TestAPIError_429(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]string{"message": "rate limited"})
	})

	_, err := client.Map(context.Background(), &MapRequest{URL: "https://x.com"})
	if !IsRateLimited(err) {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestAPIError_500_NoBody(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	_, err := client.Scrape(context.Background(), &ScrapeRequest{URL: "https://x.com"})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	// Should fall back to http.StatusText when body is empty/not JSON.
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestAPIError_ErrorFieldFallback(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request body"})
	})

	_, err := client.Scrape(context.Background(), &ScrapeRequest{URL: "https://x.com"})
	apiErr := err.(*APIError)
	if apiErr.Message != "bad request body" {
		t.Errorf("Message = %q, want 'bad request body'", apiErr.Message)
	}
}

func TestAPIError_String(t *testing.T) {
	e := &APIError{StatusCode: 403, Message: "forbidden"}
	want := "webclaw: HTTP 403 — forbidden"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// --- Helper function tests ---

func TestIsHelpers_NonAPIError(t *testing.T) {
	err := fmt.Errorf("not an api error")
	if IsAuthError(err) {
		t.Error("IsAuthError should return false for non-APIError")
	}
	if IsRateLimited(err) {
		t.Error("IsRateLimited should return false for non-APIError")
	}
	if IsNotFound(err) {
		t.Error("IsNotFound should return false for non-APIError")
	}
}

func TestIsHelpers_WrappedAPIError(t *testing.T) {
	wrapped := fmt.Errorf("call failed: %w", &APIError{StatusCode: 401})
	if !IsAuthError(wrapped) {
		t.Error("IsAuthError should unwrap and match a wrapped *APIError")
	}
	if IsNotFound(wrapped) {
		t.Error("IsNotFound should not match a wrapped 401")
	}
}

func TestAPIError_RawBodyTruncated(t *testing.T) {
	// Non-JSON body longer than the cap must not be copied verbatim.
	huge := strings.Repeat("A", 5000)
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte(huge))
	})

	_, err := client.Scrape(context.Background(), &ScrapeRequest{URL: "https://x.com"})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if len(apiErr.Message) > maxErrBodyLen+len("… (truncated)") {
		t.Errorf("Message length = %d, expected capped near %d", len(apiErr.Message), maxErrBodyLen)
	}
	if !strings.HasSuffix(apiErr.Message, "(truncated)") {
		t.Errorf("expected truncation marker, got tail %q", apiErr.Message[len(apiErr.Message)-20:])
	}
}

// --- Context cancellation ---

func TestScrape_ContextCancelled(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		json.NewEncoder(w).Encode(ScrapeResponse{})
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Scrape(ctx, &ScrapeRequest{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// --- WithHTTPClient ---

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 99 * time.Second}
	c := NewClient("k", WithHTTPClient(custom))
	if c.http != custom {
		t.Error("expected custom HTTP client to be used")
	}
}

func TestReadBodyWithLimit(t *testing.T) {
	body, err := readBodyWithLimit(strings.NewReader("1234"), 4)
	if err != nil || string(body) != "1234" {
		t.Fatalf("bounded read = %q, %v", body, err)
	}
	if _, err := readBodyWithLimit(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("expected oversized response body to fail")
	}
}

func TestRetryDelayClampsBeforeDurationConversion(t *testing.T) {
	if got := retryDelay(0, "10000000000"); got != 5*time.Second {
		t.Fatalf("oversized Retry-After delay = %s, want 5s", got)
	}
}

func TestGetRetriesTransientFailureAndEscapesID(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.EscapedPath(); got != "/v1/crawl/job%2Fwith%20space" {
			t.Errorf("escaped path = %q", got)
		}
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(CrawlStatusResponse{
			ID: "job/with space", Status: CrawlStatusCompleted,
		})
	})

	resp, err := client.GetCrawl(context.Background(), "job/with space")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != CrawlStatusCompleted || calls.Load() != 2 {
		t.Fatalf("response = %#v, calls = %d", resp, calls.Load())
	}
}

// --- BrandResponse.Decode edge case ---

func TestBrandResponse_Decode_NilData(t *testing.T) {
	resp := &BrandResponse{Data: nil}
	var dst struct{}
	if err := resp.Decode(&dst); err == nil {
		t.Error("expected error for nil data")
	}
}

// --- Example functions for godoc ---

func ExampleClient_Scrape() {
	client := NewClient("your-api-key")
	resp, err := client.Scrape(context.Background(), &ScrapeRequest{
		URL:     "https://example.com",
		Formats: []Format{FormatMarkdown},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Markdown)
}

func ExampleClient_Crawl() {
	client := NewClient("your-api-key")
	start, err := client.Crawl(context.Background(), &CrawlRequest{
		URL:      "https://example.com",
		MaxDepth: 2,
		MaxPages: 10,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.WaitForCompletion(context.Background(), start.ID, &CrawlPollOptions{
		Interval: 2 * time.Second,
		Timeout:  5 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Crawled %d pages\n", result.Completed)
}

func ExampleClient_Endpoints() {
	client := NewClient("your-api-key")
	resp, err := client.Endpoints(context.Background(), &EndpointsRequest{
		URL:               "https://example.com",
		IncludeThirdParty: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, ep := range resp.Endpoints {
		fmt.Printf("%s (%s)\n", ep.Value, ep.Kind)
	}
}

func ExampleClient_Search() {
	client := NewClient("your-api-key")
	resp, err := client.Search(context.Background(), &SearchRequest{
		Query:      "golang web scraping",
		NumResults: 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range resp.Results {
		fmt.Printf("%d. %s - %s\n", r.Position, r.Title, r.URL)
	}
}
