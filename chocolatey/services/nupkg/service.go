package nupkg

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
	"resty.dev/v3"
)

// Nupkg provides the ability to download and inspect .nupkg archives from a
// Chocolatey NuGet v2 feed.
type Nupkg struct {
	client  client.Client
	baseURL string
}

// NewNupkg returns a new Nupkg service bound to the given client and base URL.
func NewNupkg(c client.Client, baseURL string) *Nupkg {
	return &Nupkg{client: c, baseURL: strings.TrimRight(baseURL, "/")}
}

// InspectByIDAndVersion downloads the .nupkg for the given package ID and version,
// then parses and returns its contents as a NupkgInspection.
//
// API: GET /package/{id}/{version}
func (s *Nupkg) InspectByIDAndVersion(ctx context.Context, id, version string) (*NupkgInspection, *resty.Response, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("nupkg: id is required")
	}
	if version == "" {
		return nil, nil, fmt.Errorf("nupkg: version is required")
	}

	path := fmt.Sprintf("%s/%s/%s", constants.EndpointPackageDownload, id, version)
	resp, body, err := s.client.NewRequest(ctx).GetBytes(path)
	if err != nil {
		return nil, resp, fmt.Errorf("nupkg: downloading %s@%s: %w", id, version, err)
	}

	inspection, err := inspectNupkg(body)
	if err != nil {
		return nil, resp, fmt.Errorf("nupkg: inspecting %s@%s: %w", id, version, err)
	}
	return inspection, resp, nil
}

// InspectByURL downloads the .nupkg at the given direct URL and returns its contents.
// downloadURL is typically Package.DownloadURL from the packages service.
func (s *Nupkg) InspectByURL(ctx context.Context, downloadURL string) (*NupkgInspection, *resty.Response, error) {
	if downloadURL == "" {
		return nil, nil, fmt.Errorf("nupkg: downloadURL is required")
	}

	// Strip the baseURL prefix so we pass a relative path to the client.
	path := strings.TrimPrefix(downloadURL, s.baseURL)

	resp, body, err := s.client.NewRequest(ctx).GetBytes(path)
	if err != nil {
		return nil, resp, fmt.Errorf("nupkg: downloading %s: %w", downloadURL, err)
	}

	inspection, err := inspectNupkg(body)
	if err != nil {
		return nil, resp, fmt.Errorf("nupkg: inspecting %s: %w", downloadURL, err)
	}
	return inspection, resp, nil
}

// ── internal ─────────────────────────────────────────────────────────────────

// inspectNupkg parses the raw bytes of a .nupkg (ZIP) archive.
func inspectNupkg(data []byte) (*NupkgInspection, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}

	inspection := &NupkgInspection{}

	for _, f := range zr.File {
		inspection.Files = append(inspection.Files, f.Name)

		switch {
		case strings.HasSuffix(f.Name, ".nuspec"):
			content, err := readZipEntry(f)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", f.Name, err)
			}
			nuspec, err := parseNuspec(content)
			if err != nil {
				return nil, fmt.Errorf("parsing nuspec: %w", err)
			}
			inspection.Nuspec = nuspec

		case strings.EqualFold(f.Name, "tools/chocolateyInstall.ps1"):
			content, err := readZipEntry(f)
			if err != nil {
				return nil, fmt.Errorf("reading install script: %w", err)
			}
			inspection.InstallScript = parseInstallScript(string(content))

		case strings.EqualFold(f.Name, "legal/verification.txt"):
			content, err := readZipEntry(f)
			if err != nil {
				return nil, fmt.Errorf("reading verification.txt: %w", err)
			}
			inspection.Verification = parseVerification(string(content))
		}
	}

	inspection.InstallerSource = classifySource(inspection)
	return inspection, nil
}

// classifySource determines InstallerSource from the inspection results.
func classifySource(i *NupkgInspection) InstallerSource {
	if i.InstallScript == nil {
		return InstallerSourceMetaPackage
	}
	if i.InstallScript.BundledInstaller || hasBundledBinaries(i.Files) {
		return InstallerSourceBundled
	}
	if i.InstallScript.URL != "" || i.InstallScript.URL64bit != "" {
		return InstallerSourceRemoteURL
	}
	return InstallerSourceUnknown
}

// hasBundledBinaries returns true when the nupkg contains installer-like binaries
// (a strong signal that the package bundles its own installer).
func hasBundledBinaries(files []string) bool {
	for _, f := range files {
		switch strings.ToLower(filepath.Ext(f)) {
		case ".exe", ".msi", ".msu":
			return true
		}
	}
	return false
}

// readZipEntry reads and returns the full contents of a single zip.File entry.
func readZipEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// APIError wraps client.APIError for callers that need nupkg-specific 404 checks.
var _ = (*client.APIError)(nil)
