package mocks

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
	"go.uber.org/zap"
	"resty.dev/v3"
)

// registeredResponse holds a pre-canned response for a single endpoint.
type registeredResponse struct {
	statusCode int
	rawBody    []byte
	errMsg     string
}

// GenericMock is a reusable test double implementing client.Client.
// It is XML-only (all Chocolatey NuGet v2 responses are Atom XML).
type GenericMock struct {
	name            string
	responses       map[string]registeredResponse // key: "METHOD:/path"
	logger          *zap.Logger
	fixtureDir      string
	LastQueryParams map[string]string
}

// NewXMLMock creates a GenericMock configured for XML responses.
// The fixture directory is automatically resolved to the "mocks/fixtures"
// subdirectory relative to the calling package.
func NewXMLMock(name string) *GenericMock {
	fixtureDir := ""
	for i := 1; i < 10; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		dir := filepath.Dir(filename)
		if filepath.Base(dir) == "mocks" {
			continue
		}
		fixtureDir = filepath.Join(dir, "mocks", "fixtures")
		break
	}

	return &GenericMock{
		name:       name,
		responses:  make(map[string]registeredResponse),
		logger:     zap.NewNop(),
		fixtureDir: fixtureDir,
	}
}

// Register registers a mock response for the given method and path.
// If fixture is non-empty, the file is loaded from the configured fixture directory.
func (m *GenericMock) Register(method, path string, statusCode int, fixture string) {
	var body []byte
	if fixture != "" {
		data, err := m.loadFixture(fixture)
		if err != nil {
			panic(fmt.Sprintf("%s: failed to load fixture %q: %v", m.name, fixture, err))
		}
		body = data
	}
	m.responses[method+":"+path] = registeredResponse{statusCode: statusCode, rawBody: body}
}

// RegisterError registers a mock error response with the given message.
func (m *GenericMock) RegisterError(method, path string, statusCode int, errMsg string) {
	if errMsg == "" {
		errMsg = fmt.Sprintf("%s: error response %d for %s %s", m.name, statusCode, method, path)
	}
	m.responses[method+":"+path] = registeredResponse{statusCode: statusCode, errMsg: errMsg}
}

// RegisterRawBody registers a mock response with raw body bytes.
func (m *GenericMock) RegisterRawBody(method, path string, statusCode int, body []byte) {
	m.responses[method+":"+path] = registeredResponse{statusCode: statusCode, rawBody: body}
}

func (m *GenericMock) loadFixture(filename string) ([]byte, error) {
	path := filepath.Join(m.fixtureDir, filename)
	return os.ReadFile(path)
}

// dispatch is the core routing logic. Query parameters are stripped from path
// before lookup so that mock registrations use bare endpoint paths (e.g.,
// "/Packages()") and match any query string variant of that endpoint.
func (m *GenericMock) dispatch(method, path string, result any) (*resty.Response, error) {
	// Strip query string for routing purposes.
	routePath := path
	if i := strings.IndexByte(path, '?'); i >= 0 {
		routePath = path[:i]
	}

	r, ok := m.responses[method+":"+routePath]
	if !ok {
		return nil, fmt.Errorf("%s: no response registered for %s %s", m.name, method, path)
	}

	headers := http.Header{"Content-Type": {constants.ApplicationAtomXML}}
	resp := NewMockResponse(r.statusCode, headers, r.rawBody)

	if r.errMsg != "" {
		return resp, &client.APIError{
			StatusCode: r.statusCode,
			Status:     http.StatusText(r.statusCode),
			Method:     method,
			Endpoint:   path,
			Message:    r.errMsg,
		}
	}

	if result != nil && len(r.rawBody) > 0 {
		if byteSlicePtr, ok := result.(*[]byte); ok {
			*byteSlicePtr = r.rawBody
		} else {
			if err := xml.Unmarshal(r.rawBody, result); err != nil {
				return resp, fmt.Errorf("%s: unmarshal into result: %w", m.name, err)
			}
		}
	}

	return resp, nil
}

// ── client.Client implementation ──────────────────────────────────────────────

// NewRequest returns a RequestBuilder backed by this mock.
func (m *GenericMock) NewRequest(ctx context.Context) *client.RequestBuilder {
	return client.NewMockRequestBuilderWithQueryCapture(ctx,
		func(method, path string, result any) (*resty.Response, error) {
			return m.dispatch(method, path, result)
		},
		&m.LastQueryParams,
	)
}

// GetLogger returns the mock's nop logger.
func (m *GenericMock) GetLogger() *zap.Logger {
	return m.logger
}
