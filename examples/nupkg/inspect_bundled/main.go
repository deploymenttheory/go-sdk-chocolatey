// inspect_bundled demonstrates inspecting a package that bundles its installer
// binary directly inside the .nupkg archive.
//
// "7zip.install" is the canonical example: the nupkg contains the 32-bit and
// 64-bit 7-Zip executables in tools/. No remote URL is needed at install time.
// The SDK classifies these as InstallerSourceBundled.
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
)

func main() {
	c, err := chocolatey.NewClient(nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	pkg, _, err := c.Packages.GetByID(ctx, "7zip.install")
	if err != nil {
		log.Fatalf("GetByID: %v", err)
	}
	fmt.Printf("Package:     %s %s\n", pkg.ID, pkg.Version)
	fmt.Printf("DownloadURL: %s\n\n", pkg.DownloadURL)

	inspection, _, err := c.Nupkg.InspectByURL(ctx, pkg.DownloadURL)
	if err != nil {
		log.Fatalf("InspectByURL: %v", err)
	}

	fmt.Printf("Source: %s\n\n", inspection.InstallerSource)

	if inspection.InstallerSource == nupkg.InstallerSourceBundled {
		fmt.Println("Bundled installer files:")
		for _, f := range inspection.Files {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".exe" || ext == ".msi" || ext == ".msu" {
				fmt.Printf("  %s\n", f)
			}
		}
	}

	fmt.Printf("\nAll files (%d):\n", len(inspection.Files))
	for _, f := range inspection.Files {
		fmt.Printf("  %s\n", f)
	}

	if s := inspection.InstallScript; s != nil {
		fmt.Printf("\nInstall script:\n")
		fmt.Printf("  FileType:         %s\n", s.FileType)
		fmt.Printf("  BundledInstaller: %v\n", s.BundledInstaller)
	}

	// VERIFICATION.txt contains the original vendor download URLs and checksums.
	// Present for bundled packages so you can verify the bundled binary against
	// the upstream source or download it directly from the vendor.
	if v := inspection.Verification; v != nil {
		fmt.Printf("\nVendor installer (from legal/VERIFICATION.txt):\n")
		if v.VendorURL != "" {
			fmt.Printf("  URL (32-bit):  %s\n", v.VendorURL)
		}
		if v.VendorURL64bit != "" {
			fmt.Printf("  URL (64-bit):  %s\n", v.VendorURL64bit)
		}
		if v.Checksum != "" {
			fmt.Printf("  Checksum32:    %s (%s)\n", v.Checksum, v.ChecksumType)
		}
		if v.Checksum64 != "" {
			fmt.Printf("  Checksum64:    %s (%s)\n", v.Checksum64, v.ChecksumType)
		}
	} else {
		fmt.Println("\nNo VERIFICATION.txt found — download the nupkg and extract installer from tools/")
	}
}
