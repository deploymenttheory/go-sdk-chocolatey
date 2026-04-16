// Package packages provides the primary query surface for Chocolatey packages.
package packages

import "time"

// Package is a fully-resolved NuGet v2 package entry from a Chocolatey repository.
// All fields are populated from the OData Atom feed response.
type Package struct {
	// Identity
	ID      string // NuGet package identifier, e.g. "googlechrome", "7zip"
	Version string // SemVer version string, e.g. "24.9.0"
	Title   string // Display name; may differ from ID (e.g. "7-Zip" vs "7zip")

	// Authorship / provenance
	Authors string // Comma-separated list of authors
	Owners  string // Package maintainer(s) on the community repository

	// Catalog metadata
	Description  string
	Summary      string
	Tags         []string // Parsed from the space-separated Tags property
	ProjectURL   string
	IconURL      string
	LicenseURL   string
	Copyright    string
	ReleaseNotes string
	Language     string

	// Version state
	IsPrerelease            bool
	IsLatestVersion         bool // True for the latest stable version
	IsAbsoluteLatestVersion bool // True for the latest including pre-release
	Listed                  bool // False for unlisted (soft-deleted) packages

	// Artifact
	DownloadURL          string // Direct .nupkg download URL: {base}/package/{id}/{version}
	PackageHash          string // Base64-encoded SHA512 of the .nupkg file
	PackageHashAlgorithm string // Always "SHA512" for NuGet v2
	PackageSize          int64  // Size of the .nupkg in bytes

	// Statistics
	DownloadCount        int // All-time download count for this package ID
	VersionDownloadCount int // Download count for this specific version
	Published            time.Time
	Created              time.Time
	LastUpdated          time.Time

	// Dependency graph
	Dependencies []Dependency

	// Chocolatey community quality signals.
	// These fields are populated only when querying community.chocolatey.org;
	// private feeds may leave them empty.
	IsApproved              bool   // approved by Chocolatey moderators
	PackageStatus           string // "Approved", "Submitted", "Rejected", etc.
	PackageTestResultStatus string // "Passing", "Failing", "Unknown", ""
	PackageScanStatus       string // "NotFlagged", "Flagged", ""

	// Additional URLs surfaced by the Chocolatey community API.
	GalleryDetailsURL string // https://community.chocolatey.org/packages/{id}/{version}
	ProjectSourceURL  string // original project source code repository URL
	PackageSourceURL  string // package scripts on GitHub (maintainer's repo)
	DocsURL           string
	BugTrackerURL     string
}

// Dependency is a single package dependency declared in the .nuspec manifest.
type Dependency struct {
	// ID is the Chocolatey package identifier of the dependency.
	ID string
	// VersionSpec is the NuGet version range specification, e.g. "[1.0,)" or "2.0.0".
	// Empty string means any version is acceptable.
	VersionSpec string
}

// FilterOptions specifies criteria for the Search service method.
type FilterOptions struct {
	// SearchTerm is passed to the /Search() OData function as the searchTerm
	// query parameter. An empty string returns all packages (with pagination).
	SearchTerm string

	// IncludePrerelease includes pre-release versions in results when true.
	// Default is false (stable versions only).
	IncludePrerelease bool

	// Limit caps the total number of results returned across all pages.
	// 0 means fetch all pages.
	Limit int
}

// SearchResponse is the paginated result of a Search call.
type SearchResponse struct {
	// TotalCount is the total number of packages found across all pages.
	TotalCount int
	Packages   []*Package
}

// VersionsResponse is the result of a ListVersions call.
type VersionsResponse struct {
	// TotalCount is the number of versions found.
	TotalCount int
	// Versions contains the package version strings, newest first.
	Versions []string
}
