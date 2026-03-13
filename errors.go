package webclaw

import "fmt"

// APIError is returned when the API responds with a non-2xx status code.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("webclaw: HTTP %d — %s", e.StatusCode, e.Message)
}

// IsAuthError returns true if the error is a 401 Unauthorized.
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsNotFound returns true if the error is a 404 Not Found.
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
