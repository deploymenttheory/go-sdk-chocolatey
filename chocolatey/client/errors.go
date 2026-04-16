package client

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// APIError represents an HTTP error response from a Chocolatey NuGet v2 feed.
type APIError struct {
	StatusCode int
	Status     string
	Endpoint   string
	Method     string
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("chocolatey API error (%d %s) at %s %s: %s",
		e.StatusCode, e.Status, e.Method, e.Endpoint, e.Message)
}

// ParseErrorResponse constructs an APIError from a non-2xx HTTP response.
// The body may be XML or plain text — it is used as the message verbatim when
// structured error parsing is not applicable (NuGet v2 errors are not JSON).
func ParseErrorResponse(body []byte, statusCode int, status, method, endpoint string, logger *zap.Logger) error {
	msg := string(body)
	if msg == "" {
		msg = defaultMessageForStatus(statusCode)
	}

	apiErr := &APIError{
		StatusCode: statusCode,
		Status:     status,
		Endpoint:   endpoint,
		Method:     method,
		Message:    msg,
	}

	logger.Error("Chocolatey API error response",
		zap.Int("status_code", statusCode),
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.String("message", msg),
	)

	return apiErr
}

func defaultMessageForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "The request could not be understood by the server due to malformed syntax."
	case http.StatusUnauthorized:
		return "Authentication required — provide a valid API key via WithAPIKey or X-NuGet-ApiKey header."
	case http.StatusForbidden:
		return "The server understood the request but refuses to authorize it."
	case http.StatusNotFound:
		return "The requested resource was not found."
	case http.StatusTooManyRequests:
		return "Rate limit exceeded — reduce request frequency."
	case http.StatusInternalServerError:
		return "The server encountered an unexpected condition."
	case http.StatusServiceUnavailable:
		return "The server is temporarily unable to handle the request."
	default:
		return "Unknown error"
	}
}

// IsNotFound reports whether err or any error in its chain is a 404 Not Found API error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized reports whether err or any error in its chain is a 401 Unauthorized API error.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsServerError reports whether err or any error in its chain is a 5xx server error.
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= http.StatusInternalServerError && apiErr.StatusCode < 600
	}
	return false
}

// IsTooManyRequests reports whether err or any error in its chain is a 429 Too Many Requests error.
func IsTooManyRequests(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}
