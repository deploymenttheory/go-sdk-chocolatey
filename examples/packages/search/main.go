// search demonstrates searching the Chocolatey community repository.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/packages"
)

func main() {
	client, err := chocolatey.NewClient(&config.Config{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple keyword search — returns the first page of results.
	fmt.Println("=== Search: '7zip' (stable only) ===")
	result, _, err := client.Packages.Search(ctx, &packages.FilterOptions{
		SearchTerm: "7zip",
		Limit:      5, // cap at 5 for this example
	})
	if err != nil {
		log.Fatalf("search error: %v", err)
	}

	for _, p := range result.Packages {
		approved := ""
		if p.IsApproved {
			approved = " [approved]"
		}
		fmt.Printf("  %-40s %-12s %s%s\n", p.ID, p.Version, p.PackageStatus, approved)
	}
	fmt.Printf("  (%d total)\n\n", result.TotalCount)

	// Example 2: Include pre-releases.
	fmt.Println("=== Search: 'notepad' (including pre-releases) ===")
	result2, _, err := client.Packages.Search(ctx, &packages.FilterOptions{
		SearchTerm:        "notepad",
		IncludePrerelease: true,
		Limit:             5,
	})
	if err != nil {
		log.Fatalf("search error: %v", err)
	}

	for _, p := range result2.Packages {
		flags := ""
		if p.IsPrerelease {
			flags += " [pre-release]"
		}
		if p.PackageTestResultStatus != "" && p.PackageTestResultStatus != "Passing" {
			flags += " [tests:" + p.PackageTestResultStatus + "]"
		}
		fmt.Printf("  %-40s %s%s\n", p.ID, p.Version, flags)
	}
	fmt.Printf("  (%d total)\n\n", result2.TotalCount)

	// Example 3: Quality signals for the first result.
	if len(result.Packages) > 0 {
		p := result.Packages[0]
		fmt.Printf("=== Quality signals: %s ===\n", p.ID)
		fmt.Printf("  Approved:    %v\n", p.IsApproved)
		fmt.Printf("  Status:      %s\n", p.PackageStatus)
		fmt.Printf("  Test result: %s\n", p.PackageTestResultStatus)
		fmt.Printf("  Scan status: %s\n", p.PackageScanStatus)
		if p.GalleryDetailsURL != "" {
			fmt.Printf("  Gallery:     %s\n", p.GalleryDetailsURL)
		}
		if p.PackageSourceURL != "" {
			fmt.Printf("  Source:      %s\n", p.PackageSourceURL)
		}
	}
}
