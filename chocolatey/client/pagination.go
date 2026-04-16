package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"resty.dev/v3"
)

// executeODataPaginated implements requestExecutor for Transport.
// It transparently fetches all OData pages, calling mergeEntries with each
// page's raw Atom XML bytes.
//
// path must already contain all OData filter/search query parameters, built
// by the service layer using safe string concatenation (not resty SetQueryParam,
// which percent-encodes $ and OData operators). $top and $skip are appended
// directly to the path string on each iteration to avoid encoding.
//
// mergeEntries must return the number of <entry> elements found in the page.
// When entryCount < pageSize, no further pages exist and iteration stops.
func (t *Transport) executeODataPaginated(
	req *resty.Request,
	path string,
	mergeEntries func(pageXML []byte) (entryCount int, err error),
) (*resty.Response, error) {
	// Snapshot per-request headers (query params are already embedded in path).
	templateHeaders := make(map[string]string)
	for k, vs := range req.Header {
		if len(vs) > 0 {
			templateHeaders[k] = vs[0]
		}
	}

	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	pageSize := DefaultPageSize
	skip := 0

	// Determine the query string separator for the first appended param.
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}

	var lastResp *resty.Response
	for {
		// Append $top and $skip directly to avoid resty percent-encoding $ signs.
		pagePath := path + sep + "$top=" + strconv.Itoa(pageSize) + "&$skip=" + strconv.Itoa(skip)

		pageReq := t.client.R().
			SetContext(ctx).
			SetResponseBodyUnlimitedReads(true)
		for k, v := range templateHeaders {
			if v != "" {
				pageReq.SetHeader(k, v)
			}
		}

		resp, err := t.executeRequest(pageReq, "GET", pagePath)
		lastResp = resp
		if err != nil {
			return lastResp, err
		}

		n, err := mergeEntries(resp.Bytes())
		if err != nil {
			return lastResp, fmt.Errorf("merge entries: %w", err)
		}

		if n < pageSize {
			break
		}
		skip += n
	}

	return lastResp, nil
}
