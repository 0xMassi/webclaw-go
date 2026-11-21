package webclaw

import (
	"errors"
	"fmt"
)

// APIError is returned when the API responds with a non-2xx status code.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	// RequestID is the server request id from the X-Request-Id header or the
	// error body's request_id. Empty when the API returned neither.
	RequestID string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("webclaw: HTTP %d — %s (request id: %s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("webclaw: HTTP %d — %s", e.StatusCode, e.Message)
}

// IsAuthError returns true if the error is a 401 Unauthorized.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsNotFound returns true if the error is a 404 Not Found.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsForbidden returns true if the error is a 403 Forbidden. Paid-only
// endpoints (such as the X monitors and X audience export) return 403 for
// free or lapsed accounts.
func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 403
	}
	return false
}
