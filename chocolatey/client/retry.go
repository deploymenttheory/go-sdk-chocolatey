package client

import "resty.dev/v3"

// retryCondition is the resty AddRetryConditions callback.
// Returns true when the request should be retried.
//
// Retry rules:
//   - Only idempotent HTTP methods are retried (GET, HEAD, OPTIONS).
//     The Chocolatey SDK is read-only so no write methods are present.
//   - Transient server errors (500, 502, 503, 504) and 408 Request Timeout are retried.
//   - 429 Too Many Requests is retried — the NuGet community feed can rate-limit.
//   - Definitive client errors (4xx excluding 408 and 429) are never retried.
//   - Network-level errors (resp == nil) are retried for idempotent methods.
func retryCondition(resp *resty.Response, err error) bool {
	method := ""
	if resp != nil && resp.Request != nil {
		method = resp.Request.Method
	}

	if err != nil {
		return isIdempotentMethod(method)
	}

	if resp == nil {
		return false
	}

	code := resp.StatusCode()

	if isNonRetryableStatusCode(code) {
		return false
	}

	if !isIdempotentMethod(method) {
		return false
	}

	return isTransientStatusCode(code)
}

func isIdempotentMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func isTransientStatusCode(code int) bool {
	switch code {
	case 408, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func isNonRetryableStatusCode(code int) bool {
	switch code {
	case 400, 401, 402, 403, 404, 405, 406, 407, 409, 410,
		411, 412, 413, 414, 415, 416, 417, 422, 423, 424,
		426, 428, 431, 451:
		return true
	default:
		return false
	}
}
