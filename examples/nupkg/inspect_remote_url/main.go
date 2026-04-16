// inspect_remote_url demonstrates inspecting a package whose install script
// downloads the installer from a remote URL.
//
// "googlechrome" is a typical example: the nupkg is tiny (no embedded binary)
// and chocolateyInstall.ps1 contains the URL, checksum, and checksum type
// needed to download the actual MSI. The SDK classifies these as
// InstallerSourceRemoteURL.
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

	pkg, _, err := c.Packages.GetByID(ctx, "googlechrome")
	if err != nil {
		log.Fatalf("GetByID: %v", err)
	}
	fmt.Printf("Package:     %s %s\n", pkg.ID, pkg.Version)
	fmt.Printf("Approved:    %v  Status: %s  Tests: %s\n", pkg.IsApproved, pkg.PackageStatus, pkg.PackageTestResultStatus)
	fmt.Printf("DownloadURL: %s\n\n", pkg.DownloadURL)

	inspection, _, err := c.Nupkg.InspectByURL(ctx, pkg.DownloadURL)
	if err != nil {
		log.Fatalf("InspectByURL: %v", err)
	}

	fmt.Printf("Source: %s\n\n", inspection.InstallerSource)

	if inspection.InstallerSource == nupkg.InstallerSourceRemoteURL {
		s := inspection.InstallScript
		fmt.Println("Installer download info:")
		fmt.Printf("  FileType:       %s\n", s.FileType)
		if s.URL != "" {
			fmt.Printf("  URL (32-bit):   %s\n", s.URL)
			fmt.Printf("  Checksum:       %s (%s)\n", s.Checksum, s.ChecksumType)
		}
		if s.URL64bit != "" {
			fmt.Printf("  URL (64-bit):   %s\n", s.URL64bit)
			fmt.Printf("  Checksum64:     %s (%s)\n", s.Checksum64, s.ChecksumType64)
		}
		if s.SilentArgs != "" {
			fmt.Printf("  SilentArgs:     %s\n", s.SilentArgs)
		}
	}

	fmt.Printf("\nAll files (%d):\n", len(inspection.Files))
	for _, f := range inspection.Files {
		fmt.Printf("  %s\n", f)
	}
}
