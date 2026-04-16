// get_by_id_and_version demonstrates resolving a specific version of a Chocolatey
// package and obtaining the actual vendor installer download URL and checksum.
//
// pkg.DownloadURL is the .nupkg wrapper, not the installer. ResolveInstaller
// follows the full resolution chain to return the real .exe/.msi URL.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
)

func main() {
	client, err := chocolatey.NewClient(&config.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// ── Package metadata ──────────────────────────────────────────────────────

	pkg, _, err := client.Packages.GetByIDAndVersion(ctx, "notepadplusplus.install", "8.8.5")
	if err != nil {
		if chocolatey.IsNotFound(err) {
			fmt.Println("Package version not found.")
			return
		}
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Package:     %s %s\n", pkg.ID, pkg.Version)
	fmt.Printf("Authors:     %s\n", pkg.Authors)

	// pkg.DownloadURL is the .nupkg archive — not the actual installer binary.
	fmt.Printf("\nNupkg URL:   %s\n", pkg.DownloadURL)
	fmt.Printf("Hash:        %s (%s)\n", pkg.PackageHash, pkg.PackageHashAlgorithm)
	fmt.Printf("Size:        %d bytes\n", pkg.PackageSize)

	// Chocolatey community quality signals — present when querying community.chocolatey.org.
	fmt.Printf("\nApproved:    %v\n", pkg.IsApproved)
	fmt.Printf("Status:      %s\n", pkg.PackageStatus)
	fmt.Printf("Test result: %s\n", pkg.PackageTestResultStatus)
	fmt.Printf("Scan status: %s\n", pkg.PackageScanStatus)

	if pkg.GalleryDetailsURL != "" {
		fmt.Printf("\nGallery:     %s\n", pkg.GalleryDetailsURL)
	}
	if pkg.PackageSourceURL != "" {
		fmt.Printf("Source:      %s\n", pkg.PackageSourceURL)
	}

	if len(pkg.Dependencies) > 0 {
		fmt.Printf("\nDependencies (%d):\n", len(pkg.Dependencies))
		for _, d := range pkg.Dependencies {
			fmt.Printf("  %s %s\n", d.ID, d.VersionSpec)
		}
	}

	// ── Vendor installer URL ──────────────────────────────────────────────────
	//
	// ResolveInstaller returns the actual vendor .exe/.msi URL and checksum.
	// For notepadplusplus.install (a bundled package) the URL and SHA256 come
	// from legal/VERIFICATION.txt inside the nupkg.

	fmt.Println("\nResolving vendor installer URL...")
	result, err := client.ResolveInstaller(ctx, pkg.ID, pkg.Version)
	if err != nil {
		log.Fatalf("ResolveInstaller: %v", err)
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
