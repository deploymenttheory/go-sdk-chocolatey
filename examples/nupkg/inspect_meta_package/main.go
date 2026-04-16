// inspect_meta_package demonstrates inspecting a meta-package nupkg.
//
// A meta-package (e.g. "7zip") contains no install script. It exists purely to
// declare dependencies on the real installer package ("7zip.install"). The SDK
// classifies these as InstallerSourceMetaPackage.
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

	// Step 1: resolve the latest version to get the DownloadURL.
	pkg, _, err := c.Packages.GetByID(ctx, "7zip")
	if err != nil {
		log.Fatalf("GetByID: %v", err)
	}
	fmt.Printf("Package:     %s %s\n", pkg.ID, pkg.Version)
	fmt.Printf("DownloadURL: %s\n\n", pkg.DownloadURL)

	// Step 2: download and inspect the nupkg.
	inspection, _, err := c.Nupkg.InspectByURL(ctx, pkg.DownloadURL)
	if err != nil {
		log.Fatalf("InspectByURL: %v", err)
	}

	fmt.Printf("Source:      %s\n", inspection.InstallerSource)
	fmt.Printf("Files (%d):\n", len(inspection.Files))
	for _, f := range inspection.Files {
		fmt.Printf("  %s\n", f)
	}

	if inspection.InstallerSource == nupkg.InstallerSourceMetaPackage {
		fmt.Printf("\nDependencies (%d):\n", len(inspection.Nuspec.Dependencies))
		for _, d := range inspection.Nuspec.Dependencies {
			fmt.Printf("  %s  %s\n", d.ID, d.Version)
		}

		// ResolveInstaller is the high-level alternative: it follows the full
		// dependency chain automatically and returns the vendor installer URL
		// and checksum directly, without manual nupkg inspection steps.
		fmt.Println("\nTip: use ResolveInstaller to follow the dep chain automatically:")
		fmt.Println("  result, err := c.ResolveInstaller(ctx, \"7zip\", \"\")")
		fmt.Println("  // result.ResolvedPackageID == \"7zip.install\"")
		fmt.Println("  // result.URL              == vendor .exe download URL")
		fmt.Println("  // result.Checksum         == SHA256 checksum")
	}
}
