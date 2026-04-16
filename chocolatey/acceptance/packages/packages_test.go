// Package packages contains acceptance tests for the packages service.
// Tests run against a live Chocolatey NuGet v2 feed. Set CHOCOLATEY_ACCEPTANCE=true
// to enable them. All tests use well-known, stable packages (7zip) to avoid
// depending on ephemeral feed state.
package packages

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	acc "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/acceptance"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/packages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wellKnownPackage is a package that has existed on the community repository for
// many years and has numerous published versions. Tests depend only on invariants
// that are stable across all known versions (non-empty ID, hash, download URL).
const wellKnownPackage = "7zip"

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestAcceptance_Packages_GetByID(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "GetByID", "Fetching latest version of %q", wellKnownPackage)

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, resp, err := acc.Client.Packages.GetByID(ctx, wellKnownPackage)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, pkg)

	assert.Equal(t, 200, resp.StatusCode())
	assert.Equal(t, wellKnownPackage, pkg.ID, "ID should match the requested package")
	assert.NotEmpty(t, pkg.Version, "Version should not be empty")
	assert.True(t, pkg.IsLatestVersion, "GetByID should return the latest version")
	assert.False(t, pkg.IsPrerelease, "Latest stable version should not be a pre-release")
	assert.NotEmpty(t, pkg.PackageHash, "PackageHash should not be empty")
	assert.Equal(t, "SHA512", pkg.PackageHashAlgorithm)
	assert.True(t, pkg.PackageSize > 0, "PackageSize should be positive")
	assert.True(t, pkg.DownloadCount > 0, "DownloadCount should be positive")
	assert.NotEmpty(t, pkg.DownloadURL, "DownloadURL should not be empty")
	assert.True(t, strings.Contains(pkg.DownloadURL, wellKnownPackage), "DownloadURL should contain the package ID")
	assert.True(t, strings.Contains(pkg.DownloadURL, pkg.Version), "DownloadURL should contain the version")
	assert.False(t, pkg.Published.IsZero(), "Published should be set")

	acc.LogTestSuccess(t, "GetByID: id=%s version=%s hash_algo=%s size=%d",
		pkg.ID, pkg.Version, pkg.PackageHashAlgorithm, pkg.PackageSize)
}

func TestAcceptance_Packages_GetByID_notFound(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "GetByID", "Expecting 404 for non-existent package")

	ctx, cancel := acc.NewContext()
	defer cancel()

	_, _, err := acc.Client.Packages.GetByID(ctx, "this-package-definitely-does-not-exist-sdk-test")

	require.Error(t, err)
	assert.True(t, chocolatey.IsNotFound(err), "expected IsNotFound error, got: %v", err)

	acc.LogTestSuccess(t, "GetByID: non-existent package correctly returns 404")
}

// ── GetByIDAndVersion ─────────────────────────────────────────────────────────

func TestAcceptance_Packages_GetByIDAndVersion(t *testing.T) {
	acc.RequireClient(t)

	// First resolve the latest version so we have a concrete version string.
	acc.LogTestStage(t, "GetByIDAndVersion", "Resolving latest version of %q", wellKnownPackage)

	resolveCtx, resolveCancel := acc.NewContext()
	defer resolveCancel()

	latest, _, err := acc.Client.Packages.GetByID(resolveCtx, wellKnownPackage)
	require.NoError(t, err, "prerequisite: GetByID must succeed before GetByIDAndVersion")

	acc.LogTestStage(t, "GetByIDAndVersion", "Fetching %q@%s explicitly", wellKnownPackage, latest.Version)

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, resp, err := acc.Client.Packages.GetByIDAndVersion(ctx, wellKnownPackage, latest.Version)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, pkg)

	assert.Equal(t, 200, resp.StatusCode())
	assert.Equal(t, wellKnownPackage, pkg.ID)
	assert.Equal(t, latest.Version, pkg.Version, "Version should match the requested version")
	assert.NotEmpty(t, pkg.PackageHash, "PackageHash should not be empty")
	assert.Equal(t, latest.PackageHash, pkg.PackageHash, "Hash should be identical to GetByID result")
	assert.Equal(t, latest.PackageSize, pkg.PackageSize, "PackageSize should match GetByID result")

	acc.LogTestSuccess(t, "GetByIDAndVersion: id=%s version=%s", pkg.ID, pkg.Version)
}

func TestAcceptance_Packages_GetByIDAndVersion_notFound(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "GetByIDAndVersion", "Expecting 404 for version 0.0.0-nonexistent")

	ctx, cancel := acc.NewContext()
	defer cancel()

	_, _, err := acc.Client.Packages.GetByIDAndVersion(ctx, wellKnownPackage, "0.0.0-nonexistent")

	require.Error(t, err)
	assert.True(t, chocolatey.IsNotFound(err), "expected IsNotFound error, got: %v", err)

	acc.LogTestSuccess(t, "GetByIDAndVersion: non-existent version correctly returns 404")
}

// ── ListVersions ──────────────────────────────────────────────────────────────

func TestAcceptance_Packages_ListVersions(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ListVersions", "Listing all versions of %q", wellKnownPackage)

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, resp, err := acc.Client.Packages.ListVersions(ctx, wellKnownPackage)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, result)

	assert.Equal(t, 200, resp.StatusCode())
	assert.True(t, result.TotalCount > 0, "TotalCount should be positive for a well-known package")
	assert.Equal(t, result.TotalCount, len(result.Versions), "TotalCount should equal len(Versions)")
	assert.True(t, len(result.Versions) >= 3, "7zip should have at least 3 published versions")

	// Every version string must be non-empty.
	for i, v := range result.Versions {
		assert.NotEmpty(t, v, "Version at index %d should not be empty", i)
	}

	acc.LogTestSuccess(t, "ListVersions: id=%s total=%d", wellKnownPackage, result.TotalCount)
}

func TestAcceptance_Packages_ListVersions_latestIsFirst(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ListVersions", "Verifying newest-first ordering for %q", wellKnownPackage)

	resolveCtx, resolveCancel := acc.NewContext()
	defer resolveCancel()

	latest, _, err := acc.Client.Packages.GetByID(resolveCtx, wellKnownPackage)
	require.NoError(t, err, "prerequisite: GetByID must succeed")

	listCtx, listCancel := acc.NewContext()
	defer listCancel()

	result, _, err := acc.Client.Packages.ListVersions(listCtx, wellKnownPackage)
	require.NoError(t, err)
	require.True(t, len(result.Versions) > 0, "expected at least one version")

	assert.Equal(t, latest.Version, result.Versions[0],
		"first entry in ListVersions should match GetByID latest version")

	acc.LogTestSuccess(t, "ListVersions ordering: first=%s matches latest=%s",
		result.Versions[0], latest.Version)
}

// ── Search ────────────────────────────────────────────────────────────────────

func TestAcceptance_Packages_Search(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "Search", "Searching for %q", wellKnownPackage)

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, resp, err := acc.Client.Packages.Search(ctx, &packages.FilterOptions{
		SearchTerm: wellKnownPackage,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, result)

	assert.Equal(t, 200, resp.StatusCode())
	assert.True(t, result.TotalCount > 0, "Search for %q should return results", wellKnownPackage)
	assert.Equal(t, result.TotalCount, len(result.Packages))

	// At least one result should be 7zip itself.
	found := false
	for _, p := range result.Packages {
		assert.NotEmpty(t, p.ID, "Package ID should not be empty")
		assert.NotEmpty(t, p.Version, "Package Version should not be empty")
		if p.ID == wellKnownPackage {
			found = true
		}
	}
	assert.True(t, found, "search results for %q should include the package itself", wellKnownPackage)

	acc.LogTestSuccess(t, "Search: term=%q total=%d", wellKnownPackage, result.TotalCount)
}

func TestAcceptance_Packages_Search_withLimit(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "Search", "Searching with Limit=2")

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, _, err := acc.Client.Packages.Search(ctx, &packages.FilterOptions{
		SearchTerm: wellKnownPackage,
		Limit:      2,
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, len(result.Packages) <= 2, "result count should not exceed the requested Limit")

	acc.LogTestSuccess(t, "Search with Limit=2: returned %d packages", len(result.Packages))
}

func TestAcceptance_Packages_Search_includePrerelease(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "Search", "Comparing prerelease vs stable-only results")

	ctx, cancel := acc.NewContext()
	defer cancel()

	stable, _, err := acc.Client.Packages.Search(ctx, &packages.FilterOptions{
		SearchTerm:        wellKnownPackage,
		IncludePrerelease: false,
	})
	require.NoError(t, err)

	ctx2, cancel2 := acc.NewContext()
	defer cancel2()

	withPre, _, err := acc.Client.Packages.Search(ctx2, &packages.FilterOptions{
		SearchTerm:        wellKnownPackage,
		IncludePrerelease: true,
	})
	require.NoError(t, err)

	// Stable-only should return ≤ count with prerelease included.
	assert.True(t, stable.TotalCount <= withPre.TotalCount,
		"stable-only (%d) should not exceed prerelease-inclusive (%d)",
		stable.TotalCount, withPre.TotalCount)

	acc.LogTestSuccess(t, "Search prerelease comparison: stable=%d with_pre=%d",
		stable.TotalCount, withPre.TotalCount)
}

// ── Validation errors ─────────────────────────────────────────────────────────

func TestAcceptance_Packages_validation_errors(t *testing.T) {
	acc.RequireClient(t)

	t.Run("GetByID_EmptyID", func(t *testing.T) {
		_, _, err := acc.Client.Packages.GetByID(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("GetByIDAndVersion_EmptyID", func(t *testing.T) {
		_, _, err := acc.Client.Packages.GetByIDAndVersion(context.Background(), "", "1.0.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("GetByIDAndVersion_EmptyVersion", func(t *testing.T) {
		_, _, err := acc.Client.Packages.GetByIDAndVersion(context.Background(), wellKnownPackage, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version is required")
	})

	t.Run("ListVersions_EmptyID", func(t *testing.T) {
		_, _, err := acc.Client.Packages.ListVersions(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})
}
