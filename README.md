# Go SDK for Chocolatey

[![Go Report Card](https://goreportcard.com/badge/github.com/deploymenttheory/go-sdk-chocolatey)](https://goreportcard.com/report/github.com/deploymenttheory/go-sdk-chocolatey)
[![GoDoc](https://pkg.go.dev/badge/github.com/deploymenttheory/go-sdk-chocolatey)](https://pkg.go.dev/github.com/deploymenttheory/go-sdk-chocolatey)
[![License](https://img.shields.io/github/license/deploymenttheory/go-sdk-chocolatey)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deploymenttheory/go-sdk-chocolatey)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/deploymenttheory/go-sdk-chocolatey)](https://github.com/deploymenttheory/go-sdk-chocolatey/releases)
![Status: Experimental](https://img.shields.io/badge/status-experimental-yellow)

A Go client library for the [Chocolatey NuGet v2 OData API](https://community.chocolatey.org/api/v2), supporting the public community repository and private/Chocolatey for Business feeds. Resolves any package identifier to the actual vendor installer download URL and checksum — following meta-package dependency chains, parsing `legal/VERIFICATION.txt` for bundled packages, and extracting URLs from `chocolateyInstall.ps1` for remote-url packages. Production-ready transport with retries, concurrency control, TLS configuration, proxy support, and structured logging.


## Quick Start

```sh
go get github.com/deploymenttheory/go-sdk-chocolatey
```

```go
import "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"

// nil config connects to the public community repository with no authentication.
c, err := chocolatey.NewClient(nil)
if err != nil {
    log.Fatal(err)
}

// Resolve any package to its vendor installer URL and checksum.
result, err := c.ResolveInstaller(ctx, "7zip", "")
// result.URL        == "https://www.7-zip.org/a/7z2500.exe"
// result.Checksum   == "abc123..."  (SHA256)
// result.NupkgURL   == "https://community.chocolatey.org/api/v2/package/7zip.install/25.0.0"
```


## Examples

The [examples/](examples/) directory contains working programs for every SDK surface:

- **[examples/resolve/](examples/resolve/)** — `ResolveInstaller`: the primary end-to-end operation. Demonstrates all three installer delivery patterns (remote-url, bundled, meta-package) in a single program.
- **[examples/packages/get_by_id/](examples/packages/get_by_id/)** — Fetch the latest version of a package by ID; resolves to the vendor installer URL.
- **[examples/packages/get_by_id_and_version/](examples/packages/get_by_id_and_version/)** — Fetch a specific pinned version; resolves to the vendor installer URL.
- **[examples/packages/search/](examples/packages/search/)** — Full-text search across the feed with quality signal output.
- **[examples/packages/list_versions/](examples/packages/list_versions/)** — List all available versions for a package.
- **[examples/nupkg/inspect_remote_url/](examples/nupkg/inspect_remote_url/)** — Inspect a remote-url package (googlechrome); URL and checksum come from `chocolateyInstall.ps1`.
- **[examples/nupkg/inspect_bundled/](examples/nupkg/inspect_bundled/)** — Inspect a bundled package (7zip.install); URL and checksum come from `legal/VERIFICATION.txt`.
- **[examples/nupkg/inspect_meta_package/](examples/nupkg/inspect_meta_package/)** — Inspect a meta-package (7zip); shows dependency chain and `ResolveInstaller` shortcut.


## Installer Resolution

`ResolveInstaller` is the primary SDK operation. Given any package ID it returns the actual vendor `.exe`/`.msi` download URL and checksum regardless of how the package delivers its binary.

Chocolatey packages use one of three installer delivery patterns:

| Pattern | Example | How the URL is found |
|---|---|---|
| **remote-url** | `googlechrome` | Extracted from `chocolateyInstall.ps1` |
| **bundled** | `7zip.install` | Extracted from `legal/VERIFICATION.txt` inside the `.nupkg` |
| **meta-package** | `7zip` | Dependency chain followed automatically (up to 3 hops) |

```go
result, err := c.ResolveInstaller(ctx, "7zip", "") // empty version = latest

fmt.Println(result.PackageID)              // "7zip"
fmt.Println(result.ResolvedPackageID)      // "7zip.install"  (followed dep chain)
fmt.Println(result.ResolvedPackageVersion) // "25.0.0"
fmt.Println(result.InstallerSource)        // "bundled"
fmt.Println(result.URL)                    // "https://www.7-zip.org/a/7z2500.exe"
fmt.Println(result.Checksum)               // "abc123..."
fmt.Println(result.ChecksumType)           // "sha256"
fmt.Println(result.NupkgURL)               // always set — the .nupkg download URL
fmt.Println(result.DependencyChain)        // ["7zip", "7zip.install"]
```


## Client Configuration

### Creating a client

```go
import "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"

// Anonymous access to the public community repository.
c, err := chocolatey.NewClient(nil)

// Authenticated access to a private feed.
c, err := chocolatey.NewClient(
    &config.Config{
        BaseURL: "https://your-internal-feed/nuget/v2",
        APIKey:  "your-api-key",
    },
)

// With functional options.
c, err := chocolatey.NewClient(nil, chocolatey.WithDebug(), chocolatey.WithTimeout(30*time.Second))
```

### Config fields

```go
import "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"

&config.Config{
    BaseURL: "https://community.chocolatey.org/api/v2", // default when empty
    APIKey:  "",                                         // leave empty for anonymous access
}
```

### Client options

All options are applied via `chocolatey.With*` functional options passed to `NewClient`.

#### Basic Configuration

```go
chocolatey.WithBaseURL("https://...")                    // Override the feed URL
chocolatey.WithAPIKey("your-api-key")                    // NuGet API key (X-NuGet-ApiKey header)
chocolatey.WithTimeout(30*time.Second)                   // Request timeout
chocolatey.WithRetryCount(3)                             // Number of retry attempts
chocolatey.WithRetryWaitTime(2*time.Second)              // Initial retry wait time
chocolatey.WithRetryMaxWaitTime(10*time.Second)          // Maximum retry wait time
chocolatey.WithTotalRetryDuration(2*time.Minute)         // Total retry budget
```

#### TLS / Security

```go
chocolatey.WithTLSClientConfig(tlsConfig)               // Custom TLS configuration
chocolatey.WithInsecureSkipVerify()                      // Skip cert verification (dev only!)
```

#### Network

```go
chocolatey.WithProxy("http://proxy:8080")               // HTTP/HTTPS/SOCKS5 proxy
chocolatey.WithTransport(customTransport)               // Custom http.RoundTripper
```

#### Headers

```go
chocolatey.WithUserAgent("MyApp/1.0")                   // Override User-Agent
chocolatey.WithGlobalHeader("X-Custom", "value")        // Add a single header to all requests
chocolatey.WithGlobalHeaders(map[string]string{...})    // Add multiple headers to all requests
```

#### Observability

```go
chocolatey.WithLogger(zapLogger)                        // Structured logging via zap
chocolatey.WithDebug()                                  // Enable full request/response debug logging
```

#### Concurrency & Rate Limiting

```go
chocolatey.WithMaxConcurrentRequests(5)                 // Cap in-flight requests
chocolatey.WithMandatoryRequestDelay(100*time.Millisecond) // Fixed pause after every request
```

#### Example: Production Configuration

```go
import (
    "time"
    "go.uber.org/zap"
    "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
    "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
)

logger, _ := zap.NewProduction()

c, err := chocolatey.NewClient(
    &config.Config{
        BaseURL: "https://your-internal-feed/nuget/v2",
        APIKey:  os.Getenv("CHOCO_API_KEY"),
    },
    chocolatey.WithTimeout(30*time.Second),
    chocolatey.WithRetryCount(3),
    chocolatey.WithRetryMaxWaitTime(10*time.Second),
    chocolatey.WithMaxConcurrentRequests(5),
    chocolatey.WithLogger(logger),
)
```


## Services

### Packages

Query the NuGet v2 OData feed for package metadata.

```go
// Latest stable version by ID.
pkg, resp, err := c.Packages.GetByID(ctx, "googlechrome")

// Specific pinned version.
pkg, resp, err := c.Packages.GetByIDAndVersion(ctx, "notepadplusplus.install", "8.8.5")

// Full-text search with pagination.
result, resp, err := c.Packages.Search(ctx, &packages.FilterOptions{
    SearchTerm:        "7zip",
    IncludePrerelease: false,
    Limit:             20,
})

// All available versions, newest first.
versions, resp, err := c.Packages.ListVersions(ctx, "7zip")
```

The `packages.Package` struct includes the full NuGet v2 OData model plus Chocolatey community quality signals:

```go
pkg.IsApproved              // bool   — approved by Chocolatey moderators
pkg.PackageStatus           // string — "Approved", "Submitted", "Rejected", etc.
pkg.PackageTestResultStatus // string — "Passing", "Failing", "Unknown"
pkg.PackageScanStatus       // string — "NotFlagged", "Flagged"
pkg.GalleryDetailsURL       // https://community.chocolatey.org/packages/{id}/{version}
pkg.PackageSourceURL        // maintainer's package scripts on GitHub
```

### Nupkg

Download and inspect `.nupkg` archives directly.

```go
// Inspect by download URL (from pkg.DownloadURL).
inspection, resp, err := c.Nupkg.InspectByURL(ctx, pkg.DownloadURL)

// Inspect by package ID and version.
inspection, resp, err := c.Nupkg.InspectByIDAndVersion(ctx, "7zip.install", "25.0.0")

inspection.InstallerSource   // InstallerSourceRemoteURL / Bundled / MetaPackage / Unknown
inspection.InstallScript     // parsed chocolateyInstall.ps1 (URL, checksum, fileType, silentArgs)
inspection.Verification      // parsed legal/VERIFICATION.txt (vendor URL, SHA256 checksum)
inspection.Nuspec            // parsed .nuspec manifest (ID, version, dependencies)
inspection.Files             // all file paths in the archive
```


## Error Handling

```go
pkg, _, err := c.Packages.GetByID(ctx, "nonexistent-package")
if err != nil {
    if chocolatey.IsNotFound(err) {
        // Package does not exist on this feed.
    }
    // Other error (network, parse, etc.)
}
```


## Acceptance Tests

Integration tests run against the live community repository. Enable them with the `CHOCOLATEY_ACCEPTANCE=true` environment variable:

```sh
CHOCOLATEY_ACCEPTANCE=true go test ./chocolatey/acceptance/... -v -count=1 -timeout 180s
```

Unit tests require no network access and run with:

```sh
go test ./... -run TestUnit_ -v
```


## Documentation

- [Chocolatey NuGet v2 API](https://community.chocolatey.org/api/v2)
- [GoDoc](https://pkg.go.dev/github.com/deploymenttheory/go-sdk-chocolatey)


## Contributing

Contributions are welcome. Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.


## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.


## Support

- **Issues:** [GitHub Issues](https://github.com/deploymenttheory/go-sdk-chocolatey/issues)
- **Chocolatey API docs:** [community.chocolatey.org/api/v2](https://community.chocolatey.org/api/v2)


## Disclaimer

This is a community SDK and is not affiliated with or endorsed by Chocolatey Software, Inc.
