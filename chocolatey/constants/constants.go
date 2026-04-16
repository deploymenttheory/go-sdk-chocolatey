// Package constants holds SDK-wide constants for endpoints, MIME types, and defaults.
package constants

import "time"

// SDK identity.
const (
	UserAgentBase = "go-sdk-chocolatey"
	Version       = "0.1.0"
)

// DefaultBaseURL is the Chocolatey Community Repository NuGet v2 API root.
const DefaultBaseURL = "https://community.chocolatey.org/api/v2"

// HTTP client defaults.
const (
	DefaultTimeout      = 30 * time.Second
	MaxRetries          = 3
	RetryWaitTime       = 1 * time.Second
	RetryMaxWaitTime    = 10 * time.Second
	DefaultPageSize     = 50
)

// NuGet v2 OData endpoint paths (relative to BaseURL).
const (
	// EndpointPackages is the primary OData entity set. Supports $filter, $top, $skip,
	// $orderby, and semVerLevel query parameters.
	EndpointPackages = "/Packages()"

	// EndpointSearch is the full-text search function. Accepts searchTerm,
	// includePrerelease, $top, $skip, and semVerLevel query parameters.
	EndpointSearch = "/Search()"

	// EndpointFindPackagesById lists all published versions for a given package ID.
	// Accepts id and semVerLevel query parameters. Supports $orderby=Version desc.
	EndpointFindPackagesById = "/FindPackagesById()"

	// EndpointPackageDownload is the base path for direct .nupkg downloads:
	// {BaseURL}/package/{id}/{version}
	EndpointPackageDownload = "/package"
)

// MIME types used in Accept headers.
const (
	ApplicationAtomXML = "application/atom+xml"
	ApplicationXML     = "application/xml"
)
