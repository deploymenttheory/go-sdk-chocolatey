// list_versions demonstrates listing all published versions of a Chocolatey package.
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

	result, _, err := client.Packages.ListVersions(context.Background(), "7zip")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("7zip — %d published versions (newest first):\n", result.TotalCount)
	for i, v := range result.Versions {
		fmt.Printf("  [%3d] %s\n", i+1, v)
	}
}
