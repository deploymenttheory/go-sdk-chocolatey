// Package chocolatey provides the top-level entry point for the go-sdk-chocolatey SDK.
package chocolatey

import "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"

// IsNotFound reports whether err is a 404 Not Found API error.
func IsNotFound(err error) bool { return client.IsNotFound(err) }

// IsUnauthorized reports whether err is a 401 Unauthorized API error.
func IsUnauthorized(err error) bool { return client.IsUnauthorized(err) }

// IsServerError reports whether err is a 5xx server error.
func IsServerError(err error) bool { return client.IsServerError(err) }

// IsTooManyRequests reports whether err is a 429 Too Many Requests error.
func IsTooManyRequests(err error) bool { return client.IsTooManyRequests(err) }
