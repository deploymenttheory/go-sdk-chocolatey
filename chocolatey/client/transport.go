package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
	"go.uber.org/zap"
	"resty.dev/v3"
)

// Transport is the HTTP transport layer for the Chocolatey NuGet v2 API.
// It wraps a resty.Client with Chocolatey-specific behaviour: optional API key
// authentication, idempotent-only retries with exponential backoff, optional
// concurrency limiting, and structured logging.
type Transport struct {
	client        *resty.Client
	logger        *zap.Logger
	BaseURL       string
	globalHeaders map[string]string
	userAgent     string

	sem                *semaphore
	requestDelay       time.Duration
	totalRetryDuration time.Duration
}

// GetLogger returns the configured logger.
func (t *Transport) GetLogger() *zap.Logger {
	return t.logger
}

// NewTransport creates and fully configures a Chocolatey API transport.
//
// Behaviour applied at construction time:
//   - Optional API key sent as X-NuGet-ApiKey header on every request
//   - Idempotent-only retry (GET, HEAD) with exponential backoff
//   - Optional concurrency limiting via semaphore
//   - Structured request logging via zap
func NewTransport(cfg *config.Config, opts ...ClientOption) (*Transport, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	settings := &TransportSettings{
		GlobalHeaders: make(map[string]string),
	}
	for _, opt := range opts {
		if err := opt(settings); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	// Logger: caller-supplied or production default.
	logger := settings.Logger
	if logger == nil {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	// BaseURL: option > config > SDK default.
	baseURL := settings.BaseURL
	if baseURL == "" {
		baseURL = cfg.BaseURL
	}
	if baseURL == "" {
		baseURL = constants.DefaultBaseURL
	}
	baseURL = trimTrailingSlash(baseURL)

	// API key: option > config.
	apiKey := settings.APIKey
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	// UserAgent: option overrides SDK default.
	userAgent := settings.UserAgent
	if userAgent == "" {
		userAgent = fmt.Sprintf("%s/%s", constants.UserAgentBase, constants.Version)
	}

	// Timeouts/retries: option value if non-zero, else SDK default.
	timeout := settings.Timeout
	if timeout == 0 {
		timeout = constants.DefaultTimeout
	}
	retryCount := settings.RetryCount
	if retryCount == 0 {
		retryCount = constants.MaxRetries
	}
	retryWait := settings.RetryWaitTime
	if retryWait == 0 {
		retryWait = constants.RetryWaitTime
	}
	retryMaxWait := settings.RetryMaxWaitTime
	if retryMaxWait == 0 {
		retryMaxWait = constants.RetryMaxWaitTime
	}

	restyClient := resty.New()
	restyClient.SetBaseURL(baseURL)
	restyClient.SetTimeout(timeout)
	restyClient.SetRetryCount(retryCount)
	restyClient.SetRetryWaitTime(retryWait)
	restyClient.SetRetryMaxWaitTime(retryMaxWait)
	restyClient.SetHeader("User-Agent", userAgent)

	// Optional NuGet API key for private/authenticated feeds.
	if apiKey != "" {
		restyClient.SetHeader("X-NuGet-ApiKey", apiKey)
	}

	restyClient.AddRetryConditions(retryCondition)

	if settings.Debug {
		restyClient.SetDebug(true)
	}

	if settings.InsecureSkipVerify {
		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) //nolint:gosec
	} else if settings.TLSClientConfig != nil {
		restyClient.SetTLSClientConfig(settings.TLSClientConfig)
	}

	if settings.ProxyURL != "" {
		restyClient.SetProxy(settings.ProxyURL)
	}
	if settings.HTTPTransport != nil {
		restyClient.SetTransport(settings.HTTPTransport)
	}
	for k, v := range settings.GlobalHeaders {
		restyClient.SetHeader(k, v)
	}

	var sem *semaphore
	if settings.MaxConcurrentRequests > 0 {
		sem = newSemaphore(settings.MaxConcurrentRequests)
	}

	transport := &Transport{
		client:             restyClient,
		logger:             logger,
		BaseURL:            baseURL,
		globalHeaders:      settings.GlobalHeaders,
		userAgent:          userAgent,
		sem:                sem,
		requestDelay:       settings.MandatoryRequestDelay,
		totalRetryDuration: settings.TotalRetryDuration,
	}

	logger.Info("Chocolatey SDK transport created",
		zap.String("base_url", baseURL),
		zap.Bool("api_key_set", apiKey != ""),
	)

	return transport, nil
}

// NewRequest returns a RequestBuilder for this transport.
func (t *Transport) NewRequest(ctx context.Context) *RequestBuilder {
	return &RequestBuilder{
		req:      t.client.R().SetContext(ctx).SetResponseBodyUnlimitedReads(true),
		executor: t,
	}
}

// execute implements requestExecutor for Transport.
func (t *Transport) execute(req *resty.Request, method, path string, _ any) (*resty.Response, error) {
	return t.executeRequest(req, method, path)
}

// executeGetBytes implements requestExecutor for Transport.
func (t *Transport) executeGetBytes(req *resty.Request, path string) (*resty.Response, []byte, error) {
	resp, err := t.executeRequest(req, "GET", path)
	if err != nil {
		return resp, nil, err
	}
	return resp, resp.Bytes(), nil
}

// executeRequest is the central request executor. It applies the concurrency
// semaphore, total-retry deadline, mandatory per-request delay, and logging.
func (t *Transport) executeRequest(req *resty.Request, method, path string) (*resty.Response, error) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if t.totalRetryDuration > 0 {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, t.totalRetryDuration)
			defer cancel()
			req.SetContext(ctx)
		}
	}

	if t.sem != nil {
		if err := t.sem.acquire(ctx); err != nil {
			return nil, fmt.Errorf("concurrency limit: %w", err)
		}
		defer t.sem.release()
	}

	t.logger.Debug("Executing Chocolatey API request",
		zap.String("method", method),
		zap.String("path", path),
	)

	resp, execErr := req.Execute(method, path)

	if execErr != nil {
		t.logger.Error("Request failed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Error(execErr),
		)
		return resp, fmt.Errorf("request failed: %w", execErr)
	}

	if resp.IsError() {
		return resp, ParseErrorResponse(
			resp.Bytes(),
			resp.StatusCode(),
			resp.Status(),
			method,
			path,
			t.logger,
		)
	}

	t.logger.Debug("Request completed",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", resp.StatusCode()),
		zap.Duration("duration", resp.Duration()),
	)

	if t.requestDelay > 0 {
		time.Sleep(t.requestDelay)
	}

	return resp, nil
}

func trimTrailingSlash(s string) string {
	if len(s) > 0 && s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}
