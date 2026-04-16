// Package mocks provides test doubles for the Chocolatey SDK client layer.
package mocks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"resty.dev/v3"
)

// NewMockResponse creates a resty.Response for testing purposes.
func NewMockResponse(statusCode int, headers http.Header, body []byte) *resty.Response {
	if headers == nil {
		headers = make(http.Header)
	}
	if body == nil {
		body = []byte{}
	}

	status := http.StatusText(statusCode)
	if status == "" {
		status = fmt.Sprintf("%d", statusCode)
	}

	bodyReader := io.NopCloser(bytes.NewReader(body))

	req := &resty.Request{
		URL:                "",
		DoNotParseResponse: false,
	}

	resp := &resty.Response{
		Request: req,
		Body:    io.NopCloser(bytes.NewReader(body)),
		RawResponse: &http.Response{
			StatusCode: statusCode,
			Status:     status,
			Header:     headers,
			Body:       bodyReader,
		},
	}

	return resp
}
