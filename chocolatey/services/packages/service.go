package packages

import (
	"context"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
	"resty.dev/v3"
)

// odataPath builds a URL path with OData query parameters safely encoded.
// params is a sequence of alternating key, value strings. Values are encoded
// with encodeODataValue — spaces become %20 but OData-safe characters like
// $, (, ), ' are preserved, as the NuGet v2 server requires them unencoded.
func odataPath(base string, params ...string) string {
	parts := make([]string, 0, len(params)/2)
	for i := 0; i+1 < len(params); i += 2 {
		k, v := params[i], params[i+1]
		if v != "" {
			parts = append(parts, k+"="+encodeODataValue(v))
		}
	}
	if len(parts) == 0 {
		return base
	}
	return base + "?" + strings.Join(parts, "&")
}

// encodeODataValue encodes a value for an OData query parameter.
// Only spaces are encoded (as %20); OData operators $, (, ), ' are preserved.
func encodeODataValue(s string) string {
	return strings.ReplaceAll(s, " ", "%20")
}

// Packages provides read access to a Chocolatey NuGet v2 package feed.
type Packages struct {
	client  client.Client
	baseURL string // used to construct DownloadURL for each resolved package
}

// NewPackages returns a new Packages service bound to the given client and base URL.
func NewPackages(c client.Client, baseURL string) *Packages {
	return &Packages{client: c, baseURL: strings.TrimRight(baseURL, "/")}
}

// GetByID returns the latest stable version of a package.
//
// API: GET /Packages()?$filter=(tolower(Id) eq '{id}') and IsLatestVersion eq true
//
// Returns a 404 APIError when the package is not found.
func (s *Packages) GetByID(ctx context.Context, id string) (*Package, *resty.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("packages: id is required")
	}

	filter := fmt.Sprintf("(tolower(Id) eq '%s') and IsLatestVersion eq true", strings.ToLower(id))
	path := odataPath(constants.EndpointPackages,
		"$filter", filter,
		"semVerLevel", "2.0.0",
		"$top", "1",
	)

	resp, body, err := s.client.NewRequest(ctx).
		SetHeader("Accept", constants.ApplicationAtomXML).
		GetBytes(path)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: getting latest %q: %w", id, err)
	}

	entries, err := parseAtomFeed(body)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: parsing response for %q: %w", id, err)
	}

	if len(entries) == 0 {
		return nil, resp, &client.APIError{
			StatusCode: 404,
			Status:     "Not Found",
			Method:     "GET",
			Endpoint:   constants.EndpointPackages,
			Message:    fmt.Sprintf("package %q not found in repository", id),
		}
	}

	pkg, err := mapEntryToPackage(entries[0], s.baseURL)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: mapping response for %q: %w", id, err)
	}
	return pkg, resp, nil
}

// GetByIDAndVersion returns a specific version of a package.
//
// API: GET /Packages()?$filter=(tolower(Id) eq '{id}') and Version eq '{version}'
//
// Returns a 404 APIError when the package or version is not found.
func (s *Packages) GetByIDAndVersion(ctx context.Context, id, version string) (*Package, *resty.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("packages: id is required")
	}
	if version == "" {
		return nil, nil, fmt.Errorf("packages: version is required; use GetByID to resolve the latest version")
	}

	filter := fmt.Sprintf("(tolower(Id) eq '%s') and Version eq '%s'", strings.ToLower(id), version)
	path := odataPath(constants.EndpointPackages,
		"$filter", filter,
		"semVerLevel", "2.0.0",
		"$top", "1",
	)

	resp, body, err := s.client.NewRequest(ctx).
		SetHeader("Accept", constants.ApplicationAtomXML).
		GetBytes(path)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: getting %q@%s: %w", id, version, err)
	}

	entries, err := parseAtomFeed(body)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: parsing response for %q@%s: %w", id, version, err)
	}

	if len(entries) == 0 {
		return nil, resp, &client.APIError{
			StatusCode: 404,
			Status:     "Not Found",
			Method:     "GET",
			Endpoint:   constants.EndpointPackages,
			Message:    fmt.Sprintf("package %q version %q not found in repository", id, version),
		}
	}

	pkg, err := mapEntryToPackage(entries[0], s.baseURL)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: mapping response for %q@%s: %w", id, version, err)
	}
	return pkg, resp, nil
}

// ListVersions returns all published versions of a package, newest first.
//
// API: GET /FindPackagesById()?id='{id}'&$orderby=Version desc, paginated via $top/$skip.
func (s *Packages) ListVersions(ctx context.Context, id string) (*VersionsResponse, *resty.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("packages: id is required")
	}

	var result VersionsResponse

	mergeEntries := func(pageXML []byte) (int, error) {
		entries, err := parseAtomFeed(pageXML)
		if err != nil {
			return 0, fmt.Errorf("parsing versions page: %w", err)
		}
		for _, e := range entries {
			if v := e.Properties.Version; v != "" {
				result.Versions = append(result.Versions, v)
			}
		}
		return len(entries), nil
	}

	basePath := odataPath(constants.EndpointFindPackagesById,
		"id", fmt.Sprintf("'%s'", id),
		"semVerLevel", "2.0.0",
		"$orderby", "Version desc",
	)

	resp, err := s.client.NewRequest(ctx).
		SetHeader("Accept", constants.ApplicationAtomXML).
		GetPaginatedOData(basePath, mergeEntries)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: listing versions for %q: %w", id, err)
	}

	result.TotalCount = len(result.Versions)
	return &result, resp, nil
}

// Search queries the Chocolatey catalog using the full-text search function.
//
// API: GET /Search()?searchTerm='{term}'&includePrerelease={bool}, paginated via $top/$skip.
// An empty SearchTerm returns all packages (subject to pagination).
func (s *Packages) Search(ctx context.Context, opts *FilterOptions) (*SearchResponse, *resty.Response, error) {
	if opts == nil {
		opts = &FilterOptions{}
	}

	var result SearchResponse
	seen := 0

	mergeEntries := func(pageXML []byte) (int, error) {
		entries, err := parseAtomFeed(pageXML)
		if err != nil {
			return 0, fmt.Errorf("parsing search page: %w", err)
		}
		for _, e := range entries {
			if opts.Limit > 0 && seen >= opts.Limit {
				// Signal end-of-pagination by returning 0: the paginator stops
				// when entryCount < pageSize, and 0 always satisfies that.
				return 0, nil
			}
			pkg, err := mapEntryToPackage(e, s.baseURL)
			if err != nil {
				return 0, fmt.Errorf("mapping package entry: %w", err)
			}
			result.Packages = append(result.Packages, pkg)
			seen++
		}
		return len(entries), nil
	}

	includePrereleaseStr := "false"
	if opts.IncludePrerelease {
		includePrereleaseStr = "true"
	}

	basePath := odataPath(constants.EndpointSearch,
		"searchTerm", fmt.Sprintf("'%s'", opts.SearchTerm),
		"includePrerelease", includePrereleaseStr,
		"semVerLevel", "2.0.0",
	)

	resp, err := s.client.NewRequest(ctx).
		SetHeader("Accept", constants.ApplicationAtomXML).
		GetPaginatedOData(basePath, mergeEntries)
	if err != nil {
		return nil, resp, fmt.Errorf("packages: search %q: %w", opts.SearchTerm, err)
	}

	result.TotalCount = len(result.Packages)
	return &result, resp, nil
}
