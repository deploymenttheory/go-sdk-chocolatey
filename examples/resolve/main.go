// resolve demonstrates the ResolveInstaller method — the primary end-to-end
// operation of the SDK.
//
// Given any Chocolatey package ID, ResolveInstaller returns the vendor installer
// download URL and checksum regardless of how the package delivers its binary:
//
//   - remote-url  (googlechrome): URL and checksum come from chocolateyInstall.ps1
//   - bundled     (7zip.install): URL and checksum come from legal/VERIFICATION.txt
//   - meta-package (7zip):        dependency chain is followed automatically to
//     reach the real installer package
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
)

func main() {
	c, err := chocolatey.NewClient(nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	packages := []struct {
		id      string
		version string // empty = latest
		note    string
	}{
		{"googlechrome", "", "remote-url: install script downloads from vendor"},
		{"7zip.install", "", "bundled: exe embedded in nupkg, URL in VERIFICATION.txt"},
		{"7zip", "", "meta-package: delegates to 7zip.install via dependency"},
	}

	for _, p := range packages {
		fmt.Printf("\n=== %s ===\n%s\n", p.id, p.note)

		result, err := c.ResolveInstaller(ctx, p.id, p.version)
		if err != nil {
			log.Printf("  error: %v", err)
			continue
		}

		fmt.Printf("  Requested:   %s %s\n", result.PackageID, result.PackageVersion)
		if result.ResolvedPackageID != result.PackageID {
			fmt.Printf("  Resolved to: %s %s  (via dep chain: %v)\n",
				result.ResolvedPackageID, result.ResolvedPackageVersion, result.DependencyChain)
		} else {
			fmt.Printf("  Version:     %s\n", result.ResolvedPackageVersion)
		}
		fmt.Printf("  Source:      %s\n", result.InstallerSource)
		fmt.Printf("  NupkgURL:    %s\n", result.NupkgURL)

		if result.URL != "" {
			fmt.Printf("\n  Installer (32-bit):\n")
			fmt.Printf("    URL:      %s\n", result.URL)
			if result.Checksum != "" {
				fmt.Printf("    Checksum: %s (%s)\n", result.Checksum, result.ChecksumType)
			}
		}
		if result.URL64bit != "" {
			fmt.Printf("\n  Installer (64-bit):\n")
			fmt.Printf("    URL:      %s\n", result.URL64bit)
			if result.Checksum64 != "" {
				fmt.Printf("    Checksum: %s (%s)\n", result.Checksum64, result.ChecksumType64)
			}
		}
		if result.InstallerSource == nupkg.InstallerSourceBundled && result.URL == "" {
			fmt.Printf("\n  No VERIFICATION.txt found — download the nupkg and extract installer from tools/\n")
		}
	}
}
