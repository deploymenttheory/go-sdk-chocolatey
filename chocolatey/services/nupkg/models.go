package nupkg

// InstallerSource describes how the package delivers its installer binary.
type InstallerSource int

const (
	// InstallerSourceUnknown means the install script could not be parsed into a
	// recognised pattern.
	InstallerSourceUnknown InstallerSource = iota

	// InstallerSourceRemoteURL means chocolateyInstall.ps1 downloads the installer
	// from a remote URL (url / url64bit fields present).
	InstallerSourceRemoteURL

	// InstallerSourceBundled means the installer binary is embedded inside the
	// .nupkg itself (file / file64 references, or .exe/.msi files in tools/).
	InstallerSourceBundled

	// InstallerSourceMetaPackage means there is no chocolateyInstall.ps1; the
	// package delegates entirely to its <dependencies>.
	InstallerSourceMetaPackage
)

// String returns a human-readable label for the source.
func (s InstallerSource) String() string {
	switch s {
	case InstallerSourceRemoteURL:
		return "remote-url"
	case InstallerSourceBundled:
		return "bundled"
	case InstallerSourceMetaPackage:
		return "meta-package"
	default:
		return "unknown"
	}
}

// NupkgInspection holds the parsed contents of a .nupkg archive.
type NupkgInspection struct {
	// Nuspec is the parsed package manifest. Always populated when the nupkg
	// contains a valid .nuspec file.
	Nuspec *Nuspec

	// InstallScript is the result of parsing tools/chocolateyInstall.ps1.
	// Nil when no install script is present (meta-packages, tool-only packages).
	InstallScript *InstallScript

	// Verification holds data parsed from legal/VERIFICATION.txt.
	// Present for bundled packages that include the file; nil otherwise.
	// It contains the original vendor download URLs and checksums that were
	// used to source the embedded installer binary.
	Verification *Verification

	// Files lists every file path contained in the nupkg archive.
	Files []string

	// InstallerSource classifies how the package delivers its installer.
	InstallerSource InstallerSource
}

// Nuspec holds the parsed content of the .nuspec XML manifest embedded in a nupkg.
type Nuspec struct {
	ID               string
	Version          string
	Title            string
	Authors          string
	Owners           string
	Description      string
	Summary          string
	ReleaseNotes     string
	Tags             string
	ProjectURL       string
	IconURL          string
	LicenseURL       string
	Copyright        string
	PackageSourceURL string
	DocsURL          string
	BugTrackerURL    string
	Dependencies     []NuspecDependency
}

// NuspecDependency is a package dependency declared in the .nuspec manifest.
type NuspecDependency struct {
	ID      string
	Version string
}

// InstallScript holds installer metadata extracted from tools/chocolateyInstall.ps1.
//
// Chocolatey install scripts are PowerShell and have no fixed machine-readable
// schema. Fields are extracted via regular expressions over the known
// $packageArgs = @{...} hashtable pattern. Fields are empty when the script uses
// patterns the parser does not recognise.
type InstallScript struct {
	// PackageName is the value of the packageName key in $packageArgs.
	PackageName string

	// FileType is the installer type ("exe", "msi", "zip", etc.).
	FileType string

	// URL is the 32-bit installer download URL (url key).
	URL string

	// URL64bit is the 64-bit installer download URL (url64bit / url64 key).
	URL64bit string

	// Checksum is the hex-encoded checksum of the 32-bit installer.
	Checksum string

	// Checksum64 is the hex-encoded checksum of the 64-bit installer.
	Checksum64 string

	// ChecksumType is the algorithm for Checksum ("sha256", "sha1", "md5").
	ChecksumType string

	// ChecksumType64 is the algorithm for Checksum64.
	ChecksumType64 string

	// SilentArgs are the command-line arguments passed to the installer.
	SilentArgs string

	// BundledInstaller is true when the script references a local file path
	// (file / file64 keys, or $toolsDir variable patterns) rather than a URL.
	BundledInstaller bool
}
