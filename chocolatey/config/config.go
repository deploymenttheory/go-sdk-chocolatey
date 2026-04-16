// Package config holds the connection configuration for the Chocolatey SDK.
package config

// Config holds the connection configuration for a Chocolatey NuGet v2 feed.
// All fields are optional — the zero value connects to the community repository
// without authentication.
type Config struct {
	// BaseURL is the NuGet v2 API root URL.
	// Defaults to https://community.chocolatey.org/api/v2 when empty.
	// Override to target an internal or Chocolatey for Business feed.
	BaseURL string

	// APIKey is the NuGet API key for authenticated (private) feeds.
	// Sent as the X-NuGet-ApiKey request header on every request.
	// Leave empty for anonymous access to the community repository.
	APIKey string
}
