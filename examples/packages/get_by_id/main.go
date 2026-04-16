// get_by_id demonstrates resolving the latest version of a Chocolatey package
// and obtaining the actual vendor installer download URL and checksum.
//
// pkg.DownloadURL is the .nupkg wrapper, not the installer. ResolveInstaller
// follows the full resolution chain (meta-package deps, VERIFICATION.txt,
// install scripts) to return the real .exe/.msi URL you would actually download.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
)

func main() {
	// NewClient with a nil config uses the public community repository with no auth.
	// Override via config.Config{BaseURL: "...", APIKey: "..."} for private feeds.
	client, err := chocolatey.NewClient(&config.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// ── Package metadata ──────────────────────────────────────────────────────

	pkg, _, err := client.Packages.GetByID(ctx, "7zip")
	if err != nil {
		if chocolatey.IsNotFound(err) {
			fmt.Println("Package not found.")
			return
		}
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Package:     %s %s\n", pkg.ID, pkg.Version)
	fmt.Printf("Authors:     %s\n", pkg.Authors)
	fmt.Printf("Description: %s\n\n", pkg.Summary)

	// pkg.DownloadURL is the .nupkg archive — not the actual installer binary.
	fmt.Printf("Nupkg URL:   %s\n", pkg.DownloadURL)
	fmt.Printf("Hash:        %s (%s)\n", pkg.PackageHash, pkg.PackageHashAlgorithm)
	fmt.Printf("Size:        %d bytes\n\n", pkg.PackageSize)

	// Chocolatey community quality signals — present when querying community.chocolatey.org.
	fmt.Printf("Approved:    %v\n", pkg.IsApproved)
	fmt.Printf("Status:      %s\n", pkg.PackageStatus)
	fmt.Printf("Test result: %s\n", pkg.PackageTestResultStatus)
	fmt.Printf("Scan status: %s\n\n", pkg.PackageScanStatus)

	if pkg.GalleryDetailsURL != "" {
		fmt.Printf("Gallery:     %s\n", pkg.GalleryDetailsURL)
	}
	if pkg.PackageSourceURL != "" {
		fmt.Printf("Source:      %s\n", pkg.PackageSourceURL)
	}
	if pkg.ProjectURL != "" {
		fmt.Printf("Project:     %s\n\n", pkg.ProjectURL)
	}

	if len(pkg.Dependencies) > 0 {
		fmt.Printf("Dependencies (%d):\n", len(pkg.Dependencies))
		for _, d := range pkg.Dependencies {
			fmt.Printf("  %s %s\n", d.ID, d.VersionSpec)
		}
		fmt.Println()
	}

	// ── Vendor installer URL ──────────────────────────────────────────────────
	//
	// ResolveInstaller returns the actual vendor .exe/.msi URL and checksum.
	// For 7zip (a meta-package) it follows the dep chain to 7zip.install and
	// reads the URL and SHA256 from legal/VERIFICATION.txt inside that nupkg.

	fmt.Println("Resolving vendor installer URL...")
	result, err := client.ResolveInstaller(ctx, pkg.ID, pkg.Version)
	if err != nil {
		log.Fatalf("ResolveInstaller: %v", err)
	}

	if len(result.DependencyChain) > 0 {
		fmt.Printf("Resolved via: %v → %s\n", result.DependencyChain, result.ResolvedPackageID)
	}
	fmt.Printf("Source type:  %s\n\n", result.InstallerSource)

	if result.URL != "" {
		fmt.Printf("Installer (32-bit):\n")
		fmt.Printf("  URL:      %s\n", result.URL)
		if result.Checksum != "" {
			fmt.Printf("  Checksum: %s (%s)\n", result.Checksum, result.ChecksumType)
		}
	}
	if result.URL64bit != "" {
		fmt.Printf("Installer (64-bit):\n")
		fmt.Printf("  URL:      %s\n", result.URL64bit)
		if result.Checksum64 != "" {
			fmt.Printf("  Checksum: %s (%s)\n", result.Checksum64, result.ChecksumType64)
		}
	}
	if result.URL == "" && result.URL64bit == "" {
		fmt.Printf("No vendor URL found — download nupkg and extract installer from tools/\n")
		fmt.Printf("  Nupkg: %s\n", result.NupkgURL)
	}
}
