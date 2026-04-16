// Package resolve contains acceptance tests for the top-level ResolveInstaller method.
// Tests run against a live Chocolatey NuGet v2 feed. Set CHOCOLATEY_ACCEPTANCE=true
// to enable them.
package resolve

import (
	"testing"

	acc "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/acceptance"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcceptance_ResolveInstaller_metaPackage resolves 7zip, which is a
// meta-package that delegates to 7zip.install. The resolver must follow the
// dependency chain and return the vendor installer URLs from VERIFICATION.txt.
func TestAcceptance_ResolveInstaller_metaPackage(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ResolveInstaller", "Resolving meta-package: 7zip → 7zip.install")

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, err := acc.Client.ResolveInstaller(ctx, "7zip", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "7zip", result.PackageID)
	assert.Equal(t, "latest", result.PackageVersion)
	assert.Equal(t, "7zip.install", result.ResolvedPackageID, "should follow dep to 7zip.install")
	assert.NotEmpty(t, result.ResolvedPackageVersion)
	assert.Equal(t, nupkg.InstallerSourceBundled, result.InstallerSource)
	assert.NotEmpty(t, result.NupkgURL, "NupkgURL must always be set")
	assert.Contains(t, result.DependencyChain, "7zip", "chain should record traversed packages")

	// VERIFICATION.txt should surface the vendor .exe URL.
	assert.NotEmpty(t, result.URL, "vendor 32-bit URL should be extracted from VERIFICATION.txt")
	assert.NotEmpty(t, result.URL64bit, "vendor 64-bit URL should be extracted")
	assert.NotEmpty(t, result.Checksum, "checksum should be present")
	assert.Equal(t, "sha256", result.ChecksumType)

	acc.LogTestSuccess(t, "ResolveInstaller meta-package: resolved=%s@%s source=%s url=%s",
		result.ResolvedPackageID, result.ResolvedPackageVersion,
		result.InstallerSource, result.URL)
}

// TestAcceptance_ResolveInstaller_bundled resolves 7zip.install directly.
// It is a bundled package; the result should come from VERIFICATION.txt.
func TestAcceptance_ResolveInstaller_bundled(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ResolveInstaller", "Resolving bundled package: 7zip.install")

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, err := acc.Client.ResolveInstaller(ctx, "7zip.install", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "7zip.install", result.PackageID)
	assert.Equal(t, "7zip.install", result.ResolvedPackageID)
	assert.Equal(t, nupkg.InstallerSourceBundled, result.InstallerSource)
	assert.NotEmpty(t, result.NupkgURL)
	assert.Empty(t, result.DependencyChain, "no deps followed for a direct bundled package")

	assert.NotEmpty(t, result.URL, "vendor URL from VERIFICATION.txt")
	assert.NotEmpty(t, result.Checksum)
	assert.Equal(t, "sha256", result.ChecksumType)

	acc.LogTestSuccess(t, "ResolveInstaller bundled: url=%s checksum=%.16s...", result.URL, result.Checksum)
}

// TestAcceptance_ResolveInstaller_remoteURL resolves googlechrome, which
// uses a remote-url install script (no bundled binary).
func TestAcceptance_ResolveInstaller_remoteURL(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ResolveInstaller", "Resolving remote-URL package: googlechrome")

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, err := acc.Client.ResolveInstaller(ctx, "googlechrome", "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "googlechrome", result.PackageID)
	assert.Equal(t, nupkg.InstallerSourceRemoteURL, result.InstallerSource)
	assert.NotEmpty(t, result.URL, "installer URL must be present for remote-url package")
	assert.NotEmpty(t, result.URL64bit)
	assert.NotEmpty(t, result.Checksum)
	assert.NotEmpty(t, result.ChecksumType)
	assert.NotEmpty(t, result.NupkgURL)
	assert.Empty(t, result.DependencyChain)

	acc.LogTestSuccess(t, "ResolveInstaller remote-url: url=%s type=%s", result.URL, result.ChecksumType)
}

// TestAcceptance_ResolveInstaller_specificVersion resolves a pinned version of
// notepadplusplus.install, a stable bundled package with consistent versioning.
func TestAcceptance_ResolveInstaller_specificVersion(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "ResolveInstaller", "Resolving specific version: notepadplusplus.install@8.8.5")

	ctx, cancel := acc.NewContext()
	defer cancel()

	result, err := acc.Client.ResolveInstaller(ctx, "notepadplusplus.install", "8.8.5")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "8.8.5", result.PackageVersion)
	assert.Equal(t, "8.8.5", result.ResolvedPackageVersion)
	assert.Equal(t, nupkg.InstallerSourceBundled, result.InstallerSource)
	assert.NotEmpty(t, result.URL, "vendor URL from VERIFICATION.txt")
	assert.NotEmpty(t, result.Checksum)

	acc.LogTestSuccess(t, "ResolveInstaller specific version: %s@%s url=%s",
		result.ResolvedPackageID, result.ResolvedPackageVersion, result.URL)
}

// TestAcceptance_ResolveInstaller_validation_errors verifies early returns
// without making real API calls.
func TestAcceptance_ResolveInstaller_validation_errors(t *testing.T) {
	acc.RequireClient(t)

	ctx, cancel := acc.NewContext()
	defer cancel()

	t.Run("EmptyID", func(t *testing.T) {
		_, err := acc.Client.ResolveInstaller(ctx, "", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})
}
