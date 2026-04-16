// Package client provides the HTTP transport layer for the Chocolatey SDK.
// Services depend on the Client interface, not the concrete Transport, which
// enables clean unit testing through GenericMock.
package client

import (
	"context"

	"go.uber.org/zap"
)

// Client is the interface service implementations depend on.
// Transport satisfies this interface; GenericMock satisfies it in tests.
type Client interface {
	// NewRequest returns a RequestBuilder for constructing a single API request.
	// Headers, query parameters, and result targets are set on the builder before
	// calling Get/GetBytes/GetPaginatedOData to execute it. Auth, retry,
	// throttling, and concurrency limiting are applied by the transport at
	// execution time.
	NewRequest(ctx context.Context) *RequestBuilder

	// GetLogger returns the configured zap logger instance.
	GetLogger() *zap.Logger
}
