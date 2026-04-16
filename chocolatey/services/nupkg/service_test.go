package nupkg

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// buildNupkg creates a minimal in-memory .nupkg archive from named file contents.
func buildNupkg(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return buf.Bytes()
}

const sampleNuspec = `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
  <metadata>
    <id>testpkg</id>
    <version>1.2.3</version>
    <title>Test Package</title>
    <authors>Test Author</authors>
    <owners>Test Owner</owners>
    <description>A test package.</description>
    <tags>test foo bar</tags>
    <dependencies>
      <dependency id="dep1" version="[1.0,)" />
      <dependency id="dep2" version="2.0.0" />
    </dependencies>
  </metadata>
</package>`

const remoteURLScript = `$ErrorActionPreference = 'Stop'
$packageArgs = @{
  packageName   = 'testpkg'
  fileType      = 'exe'
  url           = 'https://example.com/installer.exe'
  url64bit      = 'https://example.com/installer64.exe'
  checksum      = 'abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890'
  checksum64    = '1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef'
  checksumType  = 'sha256'
  checksumType64 = 'sha256'
  silentArgs    = '/S'
  validExitCodes = @(0)
}
Install-ChocolateyPackage @packageArgs`

const bundledScript = `$ErrorActionPreference = 'Stop'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageArgs = @{
  packageName    = 'testpkg.install'
  fileType       = 'exe'
  file           = "$toolsDir\installer.exe"
  file64         = "$toolsDir\installer_x64.exe"
  silentArgs     = '/S'
  validExitCodes = @(0)
}
Install-ChocolateyInstallPackage @packageArgs`

// ── parseNuspec tests ─────────────────────────────────────────────────────────

func TestUnit_parseNuspec_happyPath(t *testing.T) {
	nuspec, err := parseNuspec([]byte(sampleNuspec))
	require.NoError(t, err)

	assert.Equal(t, "testpkg", nuspec.ID)
	assert.Equal(t, "1.2.3", nuspec.Version)
	assert.Equal(t, "Test Package", nuspec.Title)
	assert.Equal(t, "Test Author", nuspec.Authors)
	assert.Equal(t, "A test package.", nuspec.Description)
	assert.Equal(t, "test foo bar", nuspec.Tags)
	require.Len(t, nuspec.Dependencies, 2)
	assert.Equal(t, "dep1", nuspec.Dependencies[0].ID)
	assert.Equal(t, "[1.0,)", nuspec.Dependencies[0].Version)
	assert.Equal(t, "dep2", nuspec.Dependencies[1].ID)
}

func TestUnit_parseNuspec_malformed_returnsError(t *testing.T) {
	_, err := parseNuspec([]byte("<not valid xml"))
	assert.Error(t, err)
}

// ── parseInstallScript tests ──────────────────────────────────────────────────

func TestUnit_parseInstallScript_remoteURL(t *testing.T) {
	s := parseInstallScript(remoteURLScript)

	assert.Equal(t, "testpkg", s.PackageName)
	assert.Equal(t, "exe", s.FileType)
	assert.Equal(t, "https://example.com/installer.exe", s.URL)
	assert.Equal(t, "https://example.com/installer64.exe", s.URL64bit)
	assert.Equal(t, "ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890", s.Checksum)
	assert.Equal(t, "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF", s.Checksum64)
	assert.Equal(t, "sha256", s.ChecksumType)
	assert.Equal(t, "sha256", s.ChecksumType64)
	assert.Equal(t, "/S", s.SilentArgs)
	assert.False(t, s.BundledInstaller)
}

func TestUnit_parseInstallScript_bundled(t *testing.T) {
	s := parseInstallScript(bundledScript)

	assert.True(t, s.BundledInstaller)
	assert.Empty(t, s.URL)
	assert.Empty(t, s.URL64bit)
}

func TestUnit_parseInstallScript_empty(t *testing.T) {
	s := parseInstallScript("")
	assert.NotNil(t, s)
	assert.Empty(t, s.URL)
	assert.False(t, s.BundledInstaller)
}

// ── inspectNupkg tests ────────────────────────────────────────────────────────

func TestUnit_inspectNupkg_remoteURL(t *testing.T) {
	data := buildNupkg(t, map[string]string{
		"testpkg.nuspec":              sampleNuspec,
		"tools/chocolateyInstall.ps1": remoteURLScript,
	})

	i, err := inspectNupkg(data)
	require.NoError(t, err)

	require.NotNil(t, i.Nuspec)
	assert.Equal(t, "testpkg", i.Nuspec.ID)
	require.NotNil(t, i.InstallScript)
	assert.Equal(t, "https://example.com/installer.exe", i.InstallScript.URL)
	assert.Equal(t, InstallerSourceRemoteURL, i.InstallerSource)
	assert.Contains(t, i.Files, "testpkg.nuspec")
	assert.Contains(t, i.Files, "tools/chocolateyInstall.ps1")
}

func TestUnit_inspectNupkg_bundled(t *testing.T) {
	data := buildNupkg(t, map[string]string{
		"testpkg.nuspec":              sampleNuspec,
		"tools/chocolateyInstall.ps1": bundledScript,
		"tools/installer.exe":         "binary",
	})

	i, err := inspectNupkg(data)
	require.NoError(t, err)

	assert.Equal(t, InstallerSourceBundled, i.InstallerSource)
	assert.True(t, i.InstallScript.BundledInstaller)
}

func TestUnit_inspectNupkg_metaPackage(t *testing.T) {
	data := buildNupkg(t, map[string]string{
		"testpkg.nuspec": sampleNuspec,
	})

	i, err := inspectNupkg(data)
	require.NoError(t, err)

	assert.Nil(t, i.InstallScript)
	assert.Equal(t, InstallerSourceMetaPackage, i.InstallerSource)
}

func TestUnit_inspectNupkg_invalidZip_returnsError(t *testing.T) {
	_, err := inspectNupkg([]byte("not a zip file"))
	assert.Error(t, err)
}

// ── parseVerification tests ───────────────────────────────────────────────────

const verification7zip = `VERIFICATION
Verification is intended to assist the Chocolatey moderators and community
in verifying that this package's contents are trustworthy.

The installer has been downloaded from their official download link listed on <http://www.7-zip.org/download.html>
and can be verified like this:

1. Download the following installers:
  32-Bit: <http://www.7-zip.org/a/7z2600.exe>
  64-Bit: <http://www.7-zip.org/a/7z2600-x64.exe>
2. You can use one of the following methods to obtain the checksum

  checksum type: sha256
  checksum32: 3B7DCD86A17A2C4DEBAE0417DD98BB7467A69184357A23F6A3EE052356219720
  checksum64: 7B67375B2B303E05D2989F23E986126EDA67435C71231FA4B0BDAEB7A619A0A6`

const verificationNpp = `VERIFICATION
1. Download the following:
  32-Bit software: <https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v8.8.5/npp.8.8.5.Installer.exe>
  64-Bit software: <https://github.com/notepad-plus-plus/notepad-plus-plus/releases/download/v8.8.5/npp.8.8.5.Installer.x64.exe>
3. The checksums should match the following:

  checksum type: sha256
  checksum32: 05ABC57952974D08FEAFA399D6FDB37945A3FD0A10F37833DD837A5788E421D5
  checksum64: C6D1E5AACBF69AA18DF4CAF1346FD69638491A5AD0085729BAE91C662D1C62BB`

func TestUnit_parseVerification_7zip(t *testing.T) {
	v := parseVerification(verification7zip)
	require.NotNil(t, v)
	assert.Equal(t, "http://www.7-zip.org/a/7z2600.exe", v.VendorURL)
	assert.Equal(t, "http://www.7-zip.org/a/7z2600-x64.exe", v.VendorURL64bit)
	assert.Equal(t, "sha256", v.ChecksumType)
	assert.Equal(t, "3B7DCD86A17A2C4DEBAE0417DD98BB7467A69184357A23F6A3EE052356219720", v.Checksum)
	assert.Equal(t, "7B67375B2B303E05D2989F23E986126EDA67435C71231FA4B0BDAEB7A619A0A6", v.Checksum64)
	assert.NotEmpty(t, v.Raw)
}

func TestUnit_parseVerification_notepadplusplus(t *testing.T) {
	v := parseVerification(verificationNpp)
	require.NotNil(t, v)
	assert.Contains(t, v.VendorURL, "npp.8.8.5.Installer.exe")
	assert.Contains(t, v.VendorURL64bit, "npp.8.8.5.Installer.x64.exe")
	assert.Equal(t, "sha256", v.ChecksumType)
	assert.Equal(t, "05ABC57952974D08FEAFA399D6FDB37945A3FD0A10F37833DD837A5788E421D5", v.Checksum)
	assert.Equal(t, "C6D1E5AACBF69AA18DF4CAF1346FD69638491A5AD0085729BAE91C662D1C62BB", v.Checksum64)
}

func TestUnit_parseVerification_empty_returnsNil(t *testing.T) {
	assert.Nil(t, parseVerification(""))
	assert.Nil(t, parseVerification("   \n  "))
}

func TestUnit_parseVerification_noData_returnsNil(t *testing.T) {
	assert.Nil(t, parseVerification("VERIFICATION\nThis package has been verified.\n"))
}

func TestUnit_inspectNupkg_bundled_hasVerification(t *testing.T) {
	data := buildNupkg(t, map[string]string{
		"testpkg.nuspec":              sampleNuspec,
		"tools/chocolateyInstall.ps1": bundledScript,
		"tools/installer.exe":         "binary",
		"legal/VERIFICATION.txt":      verification7zip,
	})

	i, err := inspectNupkg(data)
	require.NoError(t, err)

	require.NotNil(t, i.Verification)
	assert.Equal(t, "http://www.7-zip.org/a/7z2600.exe", i.Verification.VendorURL)
	assert.Equal(t, "3B7DCD86A17A2C4DEBAE0417DD98BB7467A69184357A23F6A3EE052356219720", i.Verification.Checksum)
	assert.Equal(t, InstallerSourceBundled, i.InstallerSource)
}

func TestUnit_inspectNupkg_remoteURL_noVerification(t *testing.T) {
	data := buildNupkg(t, map[string]string{
		"testpkg.nuspec":              sampleNuspec,
		"tools/chocolateyInstall.ps1": remoteURLScript,
	})

	i, err := inspectNupkg(data)
	require.NoError(t, err)
	assert.Nil(t, i.Verification)
}

// ── service tests ─────────────────────────────────────────────────────────────

func TestUnit_Nupkg_InspectByIDAndVersion_emptyID_returnsError(t *testing.T) {
	svc := NewNupkg(mocks.NewXMLMock("test"), "https://example.com/api/v2")
	_, _, err := svc.InspectByIDAndVersion(context.Background(), "", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestUnit_Nupkg_InspectByIDAndVersion_emptyVersion_returnsError(t *testing.T) {
	svc := NewNupkg(mocks.NewXMLMock("test"), "https://example.com/api/v2")
	_, _, err := svc.InspectByIDAndVersion(context.Background(), "testpkg", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version is required")
}

func TestUnit_Nupkg_InspectByIDAndVersion_happyPath(t *testing.T) {
	nupkgData := buildNupkg(t, map[string]string{
		"testpkg.nuspec":              sampleNuspec,
		"tools/chocolateyInstall.ps1": remoteURLScript,
	})

	mock := mocks.NewXMLMock("test")
	mock.RegisterRawBody("GET", "/package/testpkg/1.2.3", 200, nupkgData)

	svc := NewNupkg(mock, "https://example.com/api/v2")
	i, resp, err := svc.InspectByIDAndVersion(context.Background(), "testpkg", "1.2.3")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, i.Nuspec)
	assert.Equal(t, "testpkg", i.Nuspec.ID)
	assert.Equal(t, InstallerSourceRemoteURL, i.InstallerSource)
}

func TestUnit_Nupkg_InspectByURL_happyPath(t *testing.T) {
	nupkgData := buildNupkg(t, map[string]string{
		"testpkg.nuspec": sampleNuspec,
	})

	mock := mocks.NewXMLMock("test")
	mock.RegisterRawBody("GET", "/package/testpkg/1.2.3", 200, nupkgData)

	svc := NewNupkg(mock, "https://example.com/api/v2")
	downloadURL := "https://example.com/api/v2/package/testpkg/1.2.3"
	i, _, err := svc.InspectByURL(context.Background(), downloadURL)
	require.NoError(t, err)
	assert.Equal(t, "testpkg", i.Nuspec.ID)
	assert.Equal(t, InstallerSourceMetaPackage, i.InstallerSource)
}
