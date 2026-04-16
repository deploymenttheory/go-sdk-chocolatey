package chocolatey

import (
	"context"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
)

// maxResolveDeps is the maximum number of meta-package hops ResolveInstaller
// will follow before returning an error.
const maxResolveDeps = 3

// InstallerResolution is the complete result of resolving a Chocolatey package
// to its vendor installer download URL and integrity checksum.
//
// The resolution process handles all three Chocolatey installer delivery patterns:
//   - remote-url:    chocolateyInstall.ps1 downloads from a vendor URL
//   - bundled:       installer binary is embedded in the .nupkg; vendor URL and
//     checksum come from legal/VERIFICATION.txt
//   - meta-package:  no installer script; the package declares a dependency on
//     the real installer package, which is followed automatically
type InstallerResolution struct {
	// PackageID is the package identifier as originally requested.
	PackageID string

	// PackageVersion is the version as originally requested, or "latest" when
	// no version was specified.
	PackageVersion string

	// ResolvedPackageID is the package that actually provided the installer.
	// Differs from PackageID when a meta-package dependency chain was followed
	// (e.g. "7zip" → "7zip.install").
	ResolvedPackageID string

	// ResolvedPackageVersion is the version of the resolved package.
	ResolvedPackageVersion string

	// InstallerSource classifies how the installer is delivered.
	InstallerSource nupkg.InstallerSource

	// URL is the 32-bit vendor installer download URL.
	// Empty for bundled packages without a VERIFICATION.txt.
	URL string

	// URL64bit is the 64-bit vendor installer download URL.
	URL64bit string

	// FileType is the installer binary type ("exe", "msi", "zip").
	FileType string

	// Checksum is the hex-encoded checksum of the 32-bit installer.
	Checksum string

	// Checksum64 is the hex-encoded checksum of the 64-bit installer.
	Checksum64 string

	// ChecksumType is the hash algorithm for Checksum ("sha256", "sha1", "md5").
	ChecksumType string

	// ChecksumType64 is the hash algorithm for Checksum64.
	ChecksumType64 string

	// NupkgURL is the direct .nupkg download URL. Always populated.
	// For InstallerSourceBundled packages where no VERIFICATION.txt is present,
	// download this URL and extract the installer binary from tools/.
	NupkgURL string

	// DependencyChain records the packages traversed to reach this result.
	// Empty when the requested package provided the installer directly.
	// Example: ["7zip", "7zip.install"]
	DependencyChain []string
}

// ResolveInstaller resolves a Chocolatey package to its vendor installer URL
// and checksum, following dependency chains and inspecting .nupkg archives as
// needed.
//
// If version is empty, the latest stable version is used.
// Meta-packages are followed up to 3 dependency hops to locate the real
// installer package. For bundled packages, vendor URLs and checksums are read
// from legal/VERIFICATION.txt inside the nupkg. For remote-url packages, they
// come directly from chocolateyInstall.ps1.
func (c *Client) ResolveInstaller(ctx context.Context, id, version string) (*InstallerResolution, error) {
	if id == "" {
		return nil, fmt.Errorf("resolve: id is required")
	}

	requestedVersion := version
	if requestedVersion == "" {
		requestedVersion = "latest"
	}

	result := &InstallerResolution{
		PackageID:      id,
		PackageVersion: requestedVersion,
	}

	if err := c.resolveStep(ctx, id, version, result, 0); err != nil {
		return nil, err
	}
	return result, nil
}

// resolveStep performs one resolution hop. depth tracks recursion to enforce maxResolveDeps.
func (c *Client) resolveStep(ctx context.Context, id, version string, result *InstallerResolution, depth int) error {
	if depth > maxResolveDeps {
		return fmt.Errorf("resolve: dependency chain exceeds maximum depth (%d) at %q", maxResolveDeps, id)
	}

	// Fetch package metadata.
	var (
		downloadURL    string
		resolvedVersion string
	)
	if version == "" {
		pkg, _, err := c.Packages.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("resolve: fetching %q: %w", id, err)
		}
		downloadURL = pkg.DownloadURL
		resolvedVersion = pkg.Version
	} else {
		pkg, _, err := c.Packages.GetByIDAndVersion(ctx, id, version)
		if err != nil {
			return fmt.Errorf("resolve: fetching %q@%s: %w", id, version, err)
		}
		downloadURL = pkg.DownloadURL
		resolvedVersion = pkg.Version
	}

	// Inspect the nupkg archive.
	inspection, _, err := c.Nupkg.InspectByURL(ctx, downloadURL)
	if err != nil {
		return fmt.Errorf("resolve: inspecting %q@%s: %w", id, resolvedVersion, err)
	}

	switch inspection.InstallerSource {

	case nupkg.InstallerSourceRemoteURL:
		s := inspection.InstallScript
		result.ResolvedPackageID = id
		result.ResolvedPackageVersion = resolvedVersion
		result.InstallerSource = nupkg.InstallerSourceRemoteURL
		result.NupkgURL = downloadURL
		result.URL = s.URL
		result.URL64bit = s.URL64bit
		result.FileType = s.FileType
		result.Checksum = s.Checksum
		result.Checksum64 = s.Checksum64
		result.ChecksumType = s.ChecksumType
		result.ChecksumType64 = s.ChecksumType64

	case nupkg.InstallerSourceBundled:
		result.ResolvedPackageID = id
		result.ResolvedPackageVersion = resolvedVersion
		result.InstallerSource = nupkg.InstallerSourceBundled
		result.NupkgURL = downloadURL
		if s := inspection.InstallScript; s != nil {
			result.FileType = s.FileType
		}
		if v := inspection.Verification; v != nil {
			result.URL = v.VendorURL
			result.URL64bit = v.VendorURL64bit
			result.Checksum = v.Checksum
			result.Checksum64 = v.Checksum64
			result.ChecksumType = v.ChecksumType
			result.ChecksumType64 = v.ChecksumType
		}

	case nupkg.InstallerSourceMetaPackage:
		result.DependencyChain = append(result.DependencyChain, id)
		if inspection.Nuspec == nil || len(inspection.Nuspec.Dependencies) == 0 {
			return fmt.Errorf("resolve: %q is a meta-package with no dependencies to follow", id)
		}
		dep := inspection.Nuspec.Dependencies[0]
		depVersion := stripVersionBrackets(dep.Version)
		return c.resolveStep(ctx, dep.ID, depVersion, result, depth+1)

	default: // InstallerSourceUnknown
		result.ResolvedPackageID = id
		result.ResolvedPackageVersion = resolvedVersion
		result.InstallerSource = nupkg.InstallerSourceUnknown
		result.NupkgURL = downloadURL
	}

	return nil
}

// stripVersionBrackets converts a NuGet version range spec to a plain version
// string, or returns "" (meaning "latest") for range expressions.
//
// Examples:
//
//	"[26.0.0]"  → "26.0.0"   (exact pinned version)
//	"[1.0,)"    → ""          (range — use latest)
//	"2.0.0"     → "2.0.0"    (already plain)
//	""          → ""
func stripVersionBrackets(spec string) string {
	s := strings.TrimSpace(spec)
	if s == "" {
		return ""
	}
	// Remove leading [ or (
	s = strings.TrimLeft(s, "[(")
	// Remove trailing ] or )
	s = strings.TrimRight(s, "])")
	// If it contains a comma it's a range — can't resolve to a single version.
	if strings.ContainsAny(s, ",") {
		return ""
	}
	return strings.TrimSpace(s)
}
