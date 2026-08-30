package webclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.webclaw.io"
	defaultTimeout = 30 * time.Second

	// maxErrBodyLen caps how many bytes of a raw, unparseable upstream
	// response body are copied into an APIError message.
	maxErrBodyLen           = 512
	maxResponseBodyLen      = 64 << 20
	maxErrorResponseBodyLen = 64 << 10
	maxGETRetries           = 2
)

func pathSegment(value string) string { return url.PathEscape(value) }

// truncate returns s shortened to at most max bytes, appending an ellipsis
// marker when content was dropped.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "… (truncated)"
}

// Client communicates with the webclaw API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets the HTTP client timeout. It applies onto whatever client
// is currently set, so order relative to WithHTTPClient does not matter.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.http == nil {
			c.http = &http.Client{}
		}
		c.http.Timeout = d
	}
}

// WithHTTPClient replaces the default HTTP client entirely. A nil argument is
// ignored so the client always keeps a usable (non-nil) http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc == nil {
			return
		}
		c.http = hc
	}
}

// NewClient creates a webclaw API client.
// The apiKey is sent as a Bearer token on every request.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// do executes an HTTP request, decoding the JSON response into dst.
// It handles auth headers and API error responses.
func (c *Client) do(ctx context.Context, method, path string, body any, dst any) error {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("webclaw: marshal request: %w", err)
		}
		bodyBytes = b
	}

	for attempt := 0; ; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
		if err != nil {
			return fmt.Errorf("webclaw: build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			if method == http.MethodGet && attempt < maxGETRetries && ctx.Err() == nil {
				if err := waitForRetry(ctx, retryDelay(attempt, "")); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("webclaw: request failed: %w", err)
		}

		if method == http.MethodGet && attempt < maxGETRetries &&
			(resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			if err := waitForRetry(ctx, retryDelay(attempt, resp.Header.Get("Retry-After"))); err != nil {
				return err
			}
			continue
		}

		return decodeResponse(resp, dst)
	}
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if seconds, err := strconv.ParseUint(strings.TrimSpace(retryAfter), 10, 64); err == nil {
		// Clamp the untrusted header before converting it to time.Duration.
		// Converting a very large second count first can overflow to a negative
		// duration and accidentally turn a requested delay into an immediate retry.
		if seconds > 5 {
			return 5 * time.Second
		}
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func readBodyWithLimit(reader io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return body, nil
}

func decodeResponse(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	limit := int64(maxResponseBodyLen)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limit = maxErrorResponseBodyLen
	}
	respBody, err := readBodyWithLimit(resp.Body, limit)
	if err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &APIError{StatusCode: resp.StatusCode, Message: err.Error()}
		}
		return fmt.Errorf("webclaw: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		// Try to extract a message from the response JSON.
		var errResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && (errResp.Message != "" || errResp.Error != "") {
			if errResp.Message != "" {
				apiErr.Message = errResp.Message
			} else {
				apiErr.Message = errResp.Error
			}
		} else if len(respBody) > 0 {
			// Unmarshal succeeded but both fields were empty (e.g. body was
			// "true", "42", or an object with different field names), or
			// unmarshal failed entirely. Use the raw body as the message,
			// capped so a misbehaving upstream can't bloat the error string.
			apiErr.Message = truncate(string(respBody), maxErrBodyLen)
		} else {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		return apiErr
	}

	if dst != nil {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("webclaw: decode response: %w", err)
		}
	}
	return nil
}
