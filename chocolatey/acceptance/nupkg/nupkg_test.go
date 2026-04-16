// Package nupkg contains acceptance tests for the nupkg service.
// Tests run against a live Chocolatey NuGet v2 feed. Set CHOCOLATEY_ACCEPTANCE=true
// to enable them.
package nupkg

import (
	"testing"

	acc "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/acceptance"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcceptance_Nupkg_InspectMetaPackage inspects the 7zip meta-package, which
// has no install script and delegates to 7zip.install via a dependency.
func TestAcceptance_Nupkg_InspectMetaPackage(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "InspectByIDAndVersion", "Inspecting 7zip meta-package nupkg")

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, _, err := acc.Client.Packages.GetByID(ctx, "7zip")
	require.NoError(t, err, "GetByID 7zip")

	ctx2, cancel2 := acc.NewContext()
	defer cancel2()

	inspection, _, err := acc.Client.Nupkg.InspectByIDAndVersion(ctx2, pkg.ID, pkg.Version)
	require.NoError(t, err)

	require.NotNil(t, inspection.Nuspec)
	assert.Equal(t, "7zip", inspection.Nuspec.ID)
	assert.NotEmpty(t, inspection.Nuspec.Version)
	assert.Equal(t, nupkg.InstallerSourceMetaPackage, inspection.InstallerSource,
		"7zip is a meta-package with no install script")
	assert.Nil(t, inspection.InstallScript)
	require.NotEmpty(t, inspection.Nuspec.Dependencies, "should declare 7zip.install dependency")
	assert.Equal(t, "7zip.install", inspection.Nuspec.Dependencies[0].ID)

	acc.LogTestSuccess(t, "InspectByIDAndVersion: source=%s deps=%d",
		inspection.InstallerSource, len(inspection.Nuspec.Dependencies))
}

// TestAcceptance_Nupkg_InspectBundledPackage inspects 7zip.install, which bundles
// the installer executables directly inside the nupkg.
func TestAcceptance_Nupkg_InspectBundledPackage(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "InspectByIDAndVersion", "Inspecting 7zip.install bundled package")

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, _, err := acc.Client.Packages.GetByID(ctx, "7zip.install")
	require.NoError(t, err, "GetByID 7zip.install")

	ctx2, cancel2 := acc.NewContext()
	defer cancel2()

	inspection, _, err := acc.Client.Nupkg.InspectByIDAndVersion(ctx2, pkg.ID, pkg.Version)
	require.NoError(t, err)

	require.NotNil(t, inspection.Nuspec)
	assert.Equal(t, "7zip.install", inspection.Nuspec.ID)
	assert.Equal(t, nupkg.InstallerSourceBundled, inspection.InstallerSource,
		"7zip.install bundles exe files in the nupkg")
	require.NotNil(t, inspection.InstallScript)

	acc.LogTestSuccess(t, "InspectByIDAndVersion: source=%s files=%d",
		inspection.InstallerSource, len(inspection.Files))
}

// TestAcceptance_Nupkg_InspectRemoteURLPackage inspects googlechrome, which
// downloads its installer from a remote URL.
func TestAcceptance_Nupkg_InspectRemoteURLPackage(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "InspectByIDAndVersion", "Inspecting googlechrome (remote-URL pattern)")

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, _, err := acc.Client.Packages.GetByID(ctx, "googlechrome")
	require.NoError(t, err, "GetByID googlechrome")

	ctx2, cancel2 := acc.NewContext()
	defer cancel2()

	inspection, _, err := acc.Client.Nupkg.InspectByIDAndVersion(ctx2, pkg.ID, pkg.Version)
	require.NoError(t, err)

	require.NotNil(t, inspection.Nuspec)
	require.NotNil(t, inspection.InstallScript)
	assert.Equal(t, nupkg.InstallerSourceRemoteURL, inspection.InstallerSource)
	assert.NotEmpty(t, inspection.InstallScript.URL, "should have 32-bit installer URL")
	assert.NotEmpty(t, inspection.InstallScript.URL64bit, "should have 64-bit installer URL")
	assert.NotEmpty(t, inspection.InstallScript.Checksum, "should have checksum")
	assert.NotEmpty(t, inspection.InstallScript.ChecksumType, "should have checksum type")

	acc.LogTestSuccess(t, "InspectByIDAndVersion: source=%s url=%s checksum_type=%s",
		inspection.InstallerSource,
		inspection.InstallScript.URL,
		inspection.InstallScript.ChecksumType)
}

// TestAcceptance_Nupkg_InspectByURL uses Package.DownloadURL directly.
func TestAcceptance_Nupkg_InspectByURL(t *testing.T) {
	acc.RequireClient(t)

	acc.LogTestStage(t, "InspectByURL", "Resolving googlechrome then inspecting via DownloadURL")

	ctx, cancel := acc.NewContext()
	defer cancel()

	pkg, _, err := acc.Client.Packages.GetByID(ctx, "googlechrome")
	require.NoError(t, err)
	require.NotEmpty(t, pkg.DownloadURL, "Package.DownloadURL must be populated")

	ctx2, cancel2 := acc.NewContext()
	defer cancel2()

	inspection, _, err := acc.Client.Nupkg.InspectByURL(ctx2, pkg.DownloadURL)
	require.NoError(t, err)

	require.NotNil(t, inspection.Nuspec)
	assert.Equal(t, pkg.ID, inspection.Nuspec.ID)
	assert.Equal(t, pkg.Version, inspection.Nuspec.Version)

	acc.LogTestSuccess(t, "InspectByURL: id=%s version=%s source=%s",
		inspection.Nuspec.ID, inspection.Nuspec.Version, inspection.InstallerSource)
}

// TestAcceptance_Nupkg_validation_errors confirms validation without hitting the API.
func TestAcceptance_Nupkg_validation_errors(t *testing.T) {
	acc.RequireClient(t)

	ctx, cancel := acc.NewContext()
	defer cancel()

	t.Run("InspectByIDAndVersion_EmptyID", func(t *testing.T) {
		_, _, err := acc.Client.Nupkg.InspectByIDAndVersion(ctx, "", "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("InspectByIDAndVersion_EmptyVersion", func(t *testing.T) {
		_, _, err := acc.Client.Nupkg.InspectByIDAndVersion(ctx, "7zip", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version is required")
	})

	t.Run("InspectByURL_EmptyURL", func(t *testing.T) {
		_, _, err := acc.Client.Nupkg.InspectByURL(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "downloadURL is required")
	})
}
