package chocolatey

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"go.uber.org/zap"
)

// ClientOption is an alias for client.ClientOption, exposed at the top level
// so callers need only import the chocolatey package.
type ClientOption = client.ClientOption

// WithBaseURL overrides the NuGet v2 API base URL.
// Use this to target a private or on-premises Chocolatey feed.
func WithBaseURL(u string) ClientOption {
	return func(s *client.TransportSettings) error {
		if u == "" {
			return fmt.Errorf("WithBaseURL: url must not be empty")
		}
		s.BaseURL = u
		return nil
	}
}

// WithAPIKey sets the NuGet API key sent as the X-NuGet-ApiKey header.
// Leave unset for the public community repository.
func WithAPIKey(key string) ClientOption {
	return func(s *client.TransportSettings) error {
		s.APIKey = key
		return nil
	}
}

// WithTimeout overrides the default HTTP request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(s *client.TransportSettings) error {
		s.Timeout = d
		return nil
	}
}

// WithRetryCount overrides the default number of retry attempts.
func WithRetryCount(n int) ClientOption {
	return func(s *client.TransportSettings) error {
		s.RetryCount = n
		return nil
	}
}

// WithRetryWaitTime overrides the initial wait duration between retries.
func WithRetryWaitTime(d time.Duration) ClientOption {
	return func(s *client.TransportSettings) error {
		s.RetryWaitTime = d
		return nil
	}
}

// WithRetryMaxWaitTime overrides the maximum wait duration between retries.
func WithRetryMaxWaitTime(d time.Duration) ClientOption {
	return func(s *client.TransportSettings) error {
		s.RetryMaxWaitTime = d
		return nil
	}
}

// WithLogger replaces the default production zap logger.
func WithLogger(l *zap.Logger) ClientOption {
	return func(s *client.TransportSettings) error {
		s.Logger = l
		return nil
	}
}

// WithDebug enables resty request/response debug logging.
func WithDebug() ClientOption {
	return func(s *client.TransportSettings) error {
		s.Debug = true
		return nil
	}
}

// WithUserAgent overrides the default SDK User-Agent header value.
func WithUserAgent(ua string) ClientOption {
	return func(s *client.TransportSettings) error {
		s.UserAgent = ua
		return nil
	}
}

// WithGlobalHeader adds a single header to every outgoing request.
func WithGlobalHeader(k, v string) ClientOption {
	return func(s *client.TransportSettings) error {
		if s.GlobalHeaders == nil {
			s.GlobalHeaders = make(map[string]string)
		}
		s.GlobalHeaders[k] = v
		return nil
	}
}

// WithGlobalHeaders merges a map of headers into every outgoing request.
func WithGlobalHeaders(h map[string]string) ClientOption {
	return func(s *client.TransportSettings) error {
		if s.GlobalHeaders == nil {
			s.GlobalHeaders = make(map[string]string)
		}
		for k, v := range h {
			s.GlobalHeaders[k] = v
		}
		return nil
	}
}

// WithProxy sets an HTTP proxy URL for all requests.
func WithProxy(u string) ClientOption {
	return func(s *client.TransportSettings) error {
		s.ProxyURL = u
		return nil
	}
}

// WithTLSClientConfig sets a custom TLS configuration.
// Ignored when WithInsecureSkipVerify is also applied.
func WithTLSClientConfig(c *tls.Config) ClientOption {
	return func(s *client.TransportSettings) error {
		s.TLSClientConfig = c
		return nil
	}
}

// WithTransport replaces the default net/http transport.
func WithTransport(rt http.RoundTripper) ClientOption {
	return func(s *client.TransportSettings) error {
		s.HTTPTransport = rt
		return nil
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// Use only for development or testing against self-signed certificates.
func WithInsecureSkipVerify() ClientOption {
	return func(s *client.TransportSettings) error {
		s.InsecureSkipVerify = true
		return nil
	}
}

// WithMaxConcurrentRequests caps the number of parallel in-flight requests.
// Zero (the default) means no limit.
func WithMaxConcurrentRequests(n int) ClientOption {
	return func(s *client.TransportSettings) error {
		s.MaxConcurrentRequests = n
		return nil
	}
}

// WithMandatoryRequestDelay inserts a fixed pause after every successful request.
// Useful for rate-limiting against feeds with strict throttling policies.
func WithMandatoryRequestDelay(d time.Duration) ClientOption {
	return func(s *client.TransportSettings) error {
		s.MandatoryRequestDelay = d
		return nil
	}
}

// WithTotalRetryDuration sets a maximum wall-clock budget for a request
// including all retry attempts. Zero disables the budget.
func WithTotalRetryDuration(d time.Duration) ClientOption {
	return func(s *client.TransportSettings) error {
		s.TotalRetryDuration = d
		return nil
	}
}
