package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ClientOption is a function that mutates TransportSettings before the Transport
// is constructed. Options are applied in order and the last write wins.
type ClientOption func(*TransportSettings) error

// TransportSettings collects all optional transport configuration.
// Zero values signal "use the built-in default".
type TransportSettings struct {
	// BaseURL overrides config.Config.BaseURL when non-empty.
	BaseURL string

	// APIKey overrides config.Config.APIKey when non-empty.
	APIKey string

	// Timeout overrides the default HTTP request timeout when non-zero.
	Timeout time.Duration

	// RetryCount overrides the default retry count when non-zero.
	RetryCount int

	// RetryWaitTime overrides the default initial retry wait when non-zero.
	RetryWaitTime time.Duration

	// RetryMaxWaitTime overrides the default maximum retry wait when non-zero.
	RetryMaxWaitTime time.Duration

	// Logger replaces the default production zap logger when non-nil.
	Logger *zap.Logger

	// Debug enables resty request/response debug logging when true.
	Debug bool

	// UserAgent replaces the default SDK user-agent string when non-empty.
	UserAgent string

	// GlobalHeaders are added to every outgoing request.
	GlobalHeaders map[string]string

	// ProxyURL sets an HTTP proxy for all requests when non-empty.
	ProxyURL string

	// TLSClientConfig sets custom TLS configuration.
	// Ignored when InsecureSkipVerify is true.
	TLSClientConfig *tls.Config

	// HTTPTransport replaces the default net/http transport when non-nil.
	HTTPTransport http.RoundTripper

	// InsecureSkipVerify disables TLS certificate verification.
	// Takes precedence over TLSClientConfig. Use only for testing.
	InsecureSkipVerify bool

	// MaxConcurrentRequests caps parallel in-flight requests.
	// 0 means no limit.
	MaxConcurrentRequests int

	// MandatoryRequestDelay inserts a fixed pause after every successful request.
	MandatoryRequestDelay time.Duration

	// TotalRetryDuration sets a maximum wall-clock budget for a request including
	// all retry attempts. Zero disables the budget.
	TotalRetryDuration time.Duration
}
