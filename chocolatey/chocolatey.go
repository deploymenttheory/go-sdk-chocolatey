// Package chocolatey is the top-level entry point for the go-sdk-chocolatey SDK.
//
// It provides a single Client struct that aggregates all service namespaces and
// exposes the full NuGet v2 OData data model exposed by Chocolatey community and
// private feeds.
//
// Basic usage:
//
//	cfg := &config.Config{}
//	c, err := chocolatey.NewClient(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	pkg, _, err := c.Packages.GetByID(ctx, "7zip")
package chocolatey

import (
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/nupkg"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/packages"
)

// Client is the top-level entry point for the go-sdk-chocolatey SDK.
// It aggregates all service namespaces behind a single, configured transport.
type Client struct {
	transport *client.Transport

	// Packages provides read access to the NuGet v2 Packages feed.
	// Use it to look up individual packages, list versions, or run full-text searches.
	Packages *packages.Packages

	// Nupkg provides the ability to download and inspect .nupkg archives.
	// Use it to retrieve the actual installer URLs, checksums, and embedded binaries
	// from a package's .nupkg file.
	Nupkg *nupkg.Nupkg
}

// NewClient creates a fully configured SDK client.
//
// cfg may be nil; in that case, all defaults apply (community repository, no auth).
// Functional options from with_options.go may be passed to override any default.
//
//	c, err := chocolatey.NewClient(nil)                           // community repo, no auth
//	c, err := chocolatey.NewClient(cfg, chocolatey.WithDebug())  // debug logging
func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	transport, err := client.NewTransport(cfg, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		transport: transport,
		Packages:  packages.NewPackages(transport, transport.BaseURL),
		Nupkg:     nupkg.NewNupkg(transport, transport.BaseURL),
	}, nil
}
