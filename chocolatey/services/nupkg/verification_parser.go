package nupkg

import (
	"regexp"
	"strings"
)

// Verification holds installer metadata parsed from legal/VERIFICATION.txt.
//
// Chocolatey package maintainers are required to include a VERIFICATION.txt in
// bundled packages that documents the original vendor download URL and checksums
// used to source the embedded installer binary. This allows package consumers to
// verify that the bundled binary matches the official vendor release.
type Verification struct {
	// VendorURL is the original 32-bit installer download URL from the software vendor.
	VendorURL string

	// VendorURL64bit is the original 64-bit installer download URL.
	VendorURL64bit string

	// Checksum is the hex-encoded checksum of the 32-bit installer.
	Checksum string

	// Checksum64 is the hex-encoded checksum of the 64-bit installer.
	Checksum64 string

	// ChecksumType is the hash algorithm ("sha256", "sha1", "md5").
	// Nearly all modern Chocolatey packages use "sha256".
	ChecksumType string

	// Raw is the full unparsed content of VERIFICATION.txt, preserved for callers
	// that need to display or log the original text.
	Raw string
}

// VERIFICATION.txt regexes.
// The format is semi-freeform prose, so patterns are intentionally loose.
// All are case-insensitive and multiline.
var (
	// 32-Bit [software]: <url> or 32-Bit: <url>
	verif32Re = regexp.MustCompile(`(?im)32-?[Bb]it[^:\n]*:\s*<?([^>\s\n]+)>?`)

	// 64-Bit [software]: <url>
	verif64Re = regexp.MustCompile(`(?im)64-?[Bb]it[^:\n]*:\s*<?([^>\s\n]+)>?`)

	// checksum type: sha256
	verifTypeRe = regexp.MustCompile(`(?im)checksum\s+type:\s*(\S+)`)

	// checksum32: HEX  (must not accidentally match checksum64)
	verifCS32Re = regexp.MustCompile(`(?im)checksum32:\s*([A-Fa-f0-9]+)`)

	// checksum64: HEX
	verifCS64Re = regexp.MustCompile(`(?im)checksum64:\s*([A-Fa-f0-9]+)`)
)

// parseVerification extracts vendor URLs and checksums from a VERIFICATION.txt string.
// Returns nil when the content is empty or contains no recognisable data.
func parseVerification(content string) *Verification {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	v := &Verification{
		Raw:            content,
		VendorURL:      extractVerif(verif32Re, content),
		VendorURL64bit: extractVerif(verif64Re, content),
		ChecksumType:   strings.ToLower(extractVerif(verifTypeRe, content)),
		Checksum:       strings.ToUpper(extractVerif(verifCS32Re, content)),
		Checksum64:     strings.ToUpper(extractVerif(verifCS64Re, content)),
	}

	// Return nil when nothing useful was extracted.
	if v.VendorURL == "" && v.VendorURL64bit == "" && v.Checksum == "" && v.Checksum64 == "" {
		return nil
	}
	return v
}

// extractVerif returns the first capture group from re applied to s, or "".
func extractVerif(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}
