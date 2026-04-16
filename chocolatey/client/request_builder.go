package client

import (
	"context"

	"resty.dev/v3"
)

// requestExecutor is the execution backend for a RequestBuilder.
// Transport implements it; mockRequestExecutor provides it in tests.
type requestExecutor interface {
	execute(req *resty.Request, method, path string, result any) (*resty.Response, error)
	executeGetBytes(req *resty.Request, path string) (*resty.Response, []byte, error)
	executeODataPaginated(req *resty.Request, path string, mergeEntries func([]byte) (int, error)) (*resty.Response, error)
}

// RequestBuilder constructs a single API request. The service layer sets
// headers, query params, and the result target before calling an execute
// method. Auth, retry, concurrency limiting, and throttling are applied
// by the transport at execution time.
//
// Usage:
//
//	resp, body, err := s.client.NewRequest(ctx).
//	    SetHeader("Accept", constants.ApplicationAtomXML).
//	    SetQueryParam("$filter", filter).
//	    SetQueryParam("semVerLevel", "2.0.0").
//	    GetBytes(constants.EndpointPackages)
type RequestBuilder struct {
	req      *resty.Request
	executor requestExecutor
	result   any
}

// SetHeader sets a request-level header. Empty values are ignored.
func (b *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	if value != "" {
		b.req.SetHeader(key, value)
	}
	return b
}

// SetQueryParam adds a URL query parameter. Empty values are ignored.
func (b *RequestBuilder) SetQueryParam(key, value string) *RequestBuilder {
	if value != "" {
		b.req.SetQueryParam(key, value)
	}
	return b
}

// SetQueryParams adds multiple URL query parameters in bulk. Empty values are ignored.
func (b *RequestBuilder) SetQueryParams(params map[string]string) *RequestBuilder {
	for k, v := range params {
		if v != "" {
			b.req.SetQueryParam(k, v)
		}
	}
	return b
}

// Get executes the request as GET against path.
// The response body is unmarshaled into the result set via SetResult, if any.
func (b *RequestBuilder) Get(path string) (*resty.Response, error) {
	return b.executor.execute(b.req, "GET", path, b.result)
}

// GetBytes executes a GET request and returns the raw response bytes without
// automatic unmarshaling. Use for XML Atom feed responses.
func (b *RequestBuilder) GetBytes(path string) (*resty.Response, []byte, error) {
	return b.executor.executeGetBytes(b.req, path)
}

// GetPaginatedOData transparently fetches all OData pages.
// mergeEntries receives the raw XML bytes of each Atom feed page and returns
// the number of <entry> elements found on that page. When entryCount < pageSize,
// no further pages exist and iteration stops.
// Query parameters already set on the builder (filter, searchTerm, $orderby, etc.)
// are forwarded as base params; $top and $skip are managed internally by the transport.
func (b *RequestBuilder) GetPaginatedOData(path string, mergeEntries func([]byte) (int, error)) (*resty.Response, error) {
	return b.executor.executeODataPaginated(b.req, path, mergeEntries)
}

// ── mock executor for unit tests ──────────────────────────────────────────────

// mockRequestExecutor backs a RequestBuilder in tests, routing execution
// through a caller-supplied dispatch function instead of a real Transport.
type mockRequestExecutor struct {
	fn              func(method, path string, result any) (*resty.Response, error)
	queryParamStore *map[string]string
}

func (m *mockRequestExecutor) execute(req *resty.Request, method, path string, result any) (*resty.Response, error) {
	m.captureQueryParams(req)
	return m.fn(method, path, result)
}

func (m *mockRequestExecutor) executeGetBytes(req *resty.Request, path string) (*resty.Response, []byte, error) {
	m.captureQueryParams(req)
	resp, err := m.fn("GET", path, nil)
	if err != nil {
		return resp, nil, err
	}
	return resp, resp.Bytes(), nil
}

func (m *mockRequestExecutor) executeODataPaginated(req *resty.Request, path string, mergeEntries func([]byte) (int, error)) (*resty.Response, error) {
	m.captureQueryParams(req)
	resp, err := m.fn("GET", path, nil)
	if err != nil {
		return resp, err
	}
	body := resp.Bytes()
	if mergeEntries != nil && len(body) > 0 {
		if _, err := mergeEntries(body); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func (m *mockRequestExecutor) captureQueryParams(req *resty.Request) {
	if m.queryParamStore != nil && req != nil {
		params := make(map[string]string)
		for k, v := range req.QueryParams {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
		if len(params) > 0 {
			*m.queryParamStore = params
		}
	}
}

// NewMockRequestBuilder returns a RequestBuilder suitable for unit tests.
// The fn callback receives the HTTP method, path, and result pointer and
// returns a pre-programmed response.
func NewMockRequestBuilder(ctx context.Context, fn func(method, path string, result any) (*resty.Response, error)) *RequestBuilder {
	return &RequestBuilder{
		req:      resty.New().R().SetContext(ctx),
		executor: &mockRequestExecutor{fn: fn},
	}
}

// NewMockRequestBuilderWithQueryCapture returns a RequestBuilder that also
// captures the query parameters into the provided map pointer.
func NewMockRequestBuilderWithQueryCapture(ctx context.Context, fn func(method, path string, result any) (*resty.Response, error), queryStore *map[string]string) *RequestBuilder {
	return &RequestBuilder{
		req:      resty.New().R().SetContext(ctx),
		executor: &mockRequestExecutor{fn: fn, queryParamStore: queryStore},
	}
}
