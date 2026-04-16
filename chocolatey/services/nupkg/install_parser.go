package nupkg

import (
	"regexp"
	"strings"
)

// The Chocolatey install script ($chocolateyInstall.ps1) is PowerShell with no
// fixed schema. The dominant pattern is a $packageArgs hashtable with named
// keys.  These regexes target the specific fields we care about using
// case-insensitive, multi-line matching.

var (
	// String value extraction: key = 'value' or key = "value"
	// The value capture group stops at the matching quote.
	psURLRe          = regexp.MustCompile(`(?im)^\s*url\s*=\s*['"]([^'"]+)['"]`)
	psURL64Re        = regexp.MustCompile(`(?im)^\s*url64(?:bit)?\s*=\s*['"]([^'"]+)['"]`)
	psChecksumRe     = regexp.MustCompile(`(?im)^\s*checksum\s*=\s*['"]([^'"]+)['"]`)
	psChecksum64Re   = regexp.MustCompile(`(?im)^\s*checksum64\s*=\s*['"]([^'"]+)['"]`)
	psChecksumTypeRe = regexp.MustCompile(`(?im)^\s*checksumType\s*=\s*['"]([^'"]+)['"]`)
	psChecksumType64Re = regexp.MustCompile(`(?im)^\s*checksumType64\s*=\s*['"]([^'"]+)['"]`)
	psPackageNameRe  = regexp.MustCompile(`(?im)^\s*packageName\s*=\s*['"]([^'"]+)['"]`)
	psFileTypeRe     = regexp.MustCompile(`(?im)^\s*fileType\s*=\s*['"]([^'"]+)['"]`)
	psSilentArgsRe   = regexp.MustCompile(`(?im)^\s*silentArgs\s*=\s*['"]([^'"]*)['"]\s*$`)

	// Presence-only: file = ... (local path, may use $variables so no value capture)
	psFileRe   = regexp.MustCompile(`(?im)^\s*file\s*=\s*\S`)
	psFile64Re = regexp.MustCompile(`(?im)^\s*file64\s*=\s*\S`)

	// Standalone variable assignments outside a hashtable, e.g.:
	//   $url = 'https://...'
	//   $checksum = 'abc123'
	psStandaloneURLRe       = regexp.MustCompile(`(?im)^\s*\$url\s*=\s*['"]([^'"]+)['"]`)
	psStandaloneURL64Re     = regexp.MustCompile(`(?im)^\s*\$url64(?:bit)?\s*=\s*['"]([^'"]+)['"]`)
	psStandaloneChecksumRe  = regexp.MustCompile(`(?im)^\s*\$checksum\s*=\s*['"]([^'"]+)['"]`)
	psStandaloneChecksum64Re = regexp.MustCompile(`(?im)^\s*\$checksum64\s*=\s*['"]([^'"]+)['"]`)
)

// extractPS extracts the first capture group from a regexp match, or returns "".
func extractPS(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// parseInstallScript extracts installer metadata from a chocolateyInstall.ps1
// source string. Fields not found in the script are left as zero values.
func parseInstallScript(content string) *InstallScript {
	s := &InstallScript{
		PackageName:    extractPS(psPackageNameRe, content),
		FileType:       strings.ToLower(extractPS(psFileTypeRe, content)),
		SilentArgs:     extractPS(psSilentArgsRe, content),
		URL:            coalesce(extractPS(psURLRe, content), extractPS(psStandaloneURLRe, content)),
		URL64bit:       coalesce(extractPS(psURL64Re, content), extractPS(psStandaloneURL64Re, content)),
		Checksum:       strings.ToUpper(coalesce(extractPS(psChecksumRe, content), extractPS(psStandaloneChecksumRe, content))),
		Checksum64:     strings.ToUpper(coalesce(extractPS(psChecksum64Re, content), extractPS(psStandaloneChecksum64Re, content))),
		ChecksumType:   strings.ToLower(extractPS(psChecksumTypeRe, content)),
		ChecksumType64: strings.ToLower(extractPS(psChecksumType64Re, content)),
	}

	// Bundled installer: file / file64 keys reference local paths (not URLs).
	if (psFileRe.MatchString(content) || psFile64Re.MatchString(content)) && s.URL == "" {
		s.BundledInstaller = true
	}

	return s
}

// coalesce returns the first non-empty string.
func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
