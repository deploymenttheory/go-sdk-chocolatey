package packages

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ── Atom XML struct definitions ───────────────────────────────────────────────
//
// NuGet v2 OData responses are Atom feeds. The three XML namespaces used are:
//   Atom: http://www.w3.org/2005/Atom
//   DS:   http://schemas.microsoft.com/ado/2007/08/dataservices
//   DSM:  http://schemas.microsoft.com/ado/2007/08/dataservices/metadata

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Entries []atomEntry `xml:"http://www.w3.org/2005/Atom entry"`
}

type atomEntry struct {
	// Atom-level fields — present on the <entry> element itself.
	// The community Chocolatey API does NOT include <d:Id>, <d:Authors>, or
	// <d:Summary> inside <m:properties>; they appear here instead.
	Title      string          `xml:"http://www.w3.org/2005/Atom title"`
	Summary    string          `xml:"http://www.w3.org/2005/Atom summary"`
	Author     atomAuthor      `xml:"http://www.w3.org/2005/Atom author"`
	Content    atomContent     `xml:"http://www.w3.org/2005/Atom content"`
	Properties entryProperties `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices/metadata properties"`
}

// atomAuthor holds the <author><name> element inside an Atom entry.
type atomAuthor struct {
	Name string `xml:"http://www.w3.org/2005/Atom name"`
}

// atomContent holds the <content> element whose src attribute is the .nupkg URL.
type atomContent struct {
	Src string `xml:"src,attr"`
}

// entryProperties maps each <d:*> OData property. All values are captured as
// strings; typed conversion happens in mapEntryToPackage. Boolean properties
// carry an m:type="Edm.Boolean" XML attribute that Go's xml decoder ignores,
// but the text content ("true"/"false") is what we parse.
type entryProperties struct {
	ID                      string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Id"`
	Version                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Version"`
	Title                   string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Title"`
	Description             string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Description"`
	Summary                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Summary"`
	Authors                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Authors"`
	Owners                  string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Owners"`
	Tags                    string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Tags"`
	ProjectURL              string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices ProjectUrl"`
	IconURL                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices IconUrl"`
	LicenseURL              string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices LicenseUrl"`
	Copyright               string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Copyright"`
	ReleaseNotes            string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices ReleaseNotes"`
	Language                string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Language"`
	IsPrerelease            string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices IsPrerelease"`
	IsLatestVersion         string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices IsLatestVersion"`
	IsAbsoluteLatestVersion string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices IsAbsoluteLatestVersion"`
	Listed                  string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Listed"`
	DownloadCount           string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices DownloadCount"`
	VersionDownloadCount    string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices VersionDownloadCount"`
	PackageHash             string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageHash"`
	PackageHashAlgorithm    string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageHashAlgorithm"`
	PackageSize             string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageSize"`
	Published               string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Published"`
	Created                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Created"`
	LastUpdated             string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices LastUpdated"`
	Dependencies            string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Dependencies"`

	// Chocolatey community-specific fields.
	IsApproved              string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices IsApproved"`
	PackageStatus           string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageStatus"`
	PackageTestResultStatus string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageTestResultStatus"`
	PackageScanStatus       string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageScanStatus"`
	GalleryDetailsUrl       string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices GalleryDetailsUrl"`
	ProjectSourceUrl        string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices ProjectSourceUrl"`
	PackageSourceUrl        string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices PackageSourceUrl"`
	DocsUrl                 string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices DocsUrl"`
	BugTrackerUrl           string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices BugTrackerUrl"`
}

// ── Parsing functions ─────────────────────────────────────────────────────────

// parseAtomFeed parses raw NuGet v2 OData Atom XML bytes into a slice of entries.
//
// The Chocolatey community server occasionally returns malformed XML at the end
// of paginated responses (a <link rel="next"> element containing an ASP.NET
// error node and no closing tags). Go's strict XML decoder treats this as a
// fatal parse error. To handle this gracefully, parseAtomFeed uses a streaming
// xml.Decoder and accumulates <entry> elements one at a time, stopping cleanly
// when it encounters the malformed tail rather than returning an error.
func parseAtomFeed(data []byte) ([]atomEntry, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))

	const (
		atomNS = "http://www.w3.org/2005/Atom"
		entry  = "entry"
	)

	var entries []atomEntry
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// The Chocolatey server sometimes emits malformed XML after the
			// last </entry> (a broken <link rel="next"> element). Stop here
			// and return whatever entries were successfully parsed.
			break
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if start.Name.Space != atomNS || start.Name.Local != entry {
			continue
		}

		var e atomEntry
		if err := dec.DecodeElement(&e, &start); err != nil {
			// Malformed entry — stop iteration.
			break
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// mapEntryToPackage converts a raw atomEntry into a typed Package.
// baseURL is provided for fallback DownloadURL construction when the Atom
// <content src="..."> element is absent (e.g., in unit-test fixtures).
func mapEntryToPackage(e atomEntry, baseURL string) (*Package, error) {
	p := e.Properties

	downloadCount, _ := strconv.Atoi(p.DownloadCount)
	versionDownloadCount, _ := strconv.Atoi(p.VersionDownloadCount)
	packageSize, _ := strconv.ParseInt(p.PackageSize, 10, 64)

	published := parseODataDateTime(p.Published)
	created := parseODataDateTime(p.Created)
	lastUpdated := parseODataDateTime(p.LastUpdated)

	// Package ID: the community Chocolatey API does not include <d:Id> inside
	// <m:properties>. The canonical ID is the Atom entry <title> element.
	// Fall back to the <d:Id> property for private/test feeds that include it.
	id := strings.TrimSpace(e.Title)
	if id == "" {
		id = p.ID
	}

	// Authors: Atom entry <author><name> takes precedence over <d:Authors>.
	authors := strings.TrimSpace(e.Author.Name)
	if authors == "" {
		authors = p.Authors
	}

	// Summary: Atom entry <summary> takes precedence over <d:Summary>.
	summary := strings.TrimSpace(e.Summary)
	if summary == "" {
		summary = p.Summary
	}

	// DownloadURL: Atom entry <content src="..."> contains the direct .nupkg URL.
	// Fall back to constructing it from baseURL + id + version (e.g., fixtures).
	downloadURL := strings.TrimSpace(e.Content.Src)
	if downloadURL == "" && id != "" && p.Version != "" {
		downloadURL = fmt.Sprintf("%s/package/%s/%s", baseURL, id, p.Version)
	}

	return &Package{
		ID:                      id,
		Version:                 p.Version,
		Title:                   p.Title,
		Authors:                 authors,
		Owners:                  p.Owners,
		Description:             p.Description,
		Summary:                 summary,
		Tags:                    parseTags(p.Tags),
		ProjectURL:              p.ProjectURL,
		IconURL:                 p.IconURL,
		LicenseURL:              p.LicenseURL,
		Copyright:               p.Copyright,
		ReleaseNotes:            p.ReleaseNotes,
		Language:                p.Language,
		IsPrerelease:            parseBool(p.IsPrerelease),
		IsLatestVersion:         parseBool(p.IsLatestVersion),
		IsAbsoluteLatestVersion: parseBool(p.IsAbsoluteLatestVersion),
		Listed:                  parseBool(p.Listed),
		DownloadURL:             downloadURL,
		PackageHash:             p.PackageHash,
		PackageHashAlgorithm:    p.PackageHashAlgorithm,
		PackageSize:             packageSize,
		DownloadCount:           downloadCount,
		VersionDownloadCount:    versionDownloadCount,
		Published:               published,
		Created:                 created,
		LastUpdated:             lastUpdated,
		Dependencies:            parseDependencies(p.Dependencies),
		IsApproved:              parseBool(p.IsApproved),
		PackageStatus:           p.PackageStatus,
		PackageTestResultStatus: p.PackageTestResultStatus,
		PackageScanStatus:       p.PackageScanStatus,
		GalleryDetailsURL:       p.GalleryDetailsUrl,
		ProjectSourceURL:        p.ProjectSourceUrl,
		PackageSourceURL:        p.PackageSourceUrl,
		DocsURL:                 p.DocsUrl,
		BugTrackerURL:           p.BugTrackerUrl,
	}, nil
}

// parseDependencies parses the NuGet v2 Dependencies string.
// Format: "id1:versionSpec:targetFramework|id2:versionSpec:|id3:"
// Each pipe-separated segment has the form "id:versionSpec:targetFramework"
// where versionSpec and targetFramework may be empty. The third field
// (targetFramework) is intentionally discarded.
func parseDependencies(raw string) []Dependency {
	if raw == "" {
		return nil
	}
	var deps []Dependency
	for _, segment := range strings.Split(raw, "|") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		// Split into at most 3 parts: id, versionSpec, targetFramework.
		parts := strings.SplitN(segment, ":", 3)
		id := strings.TrimSpace(parts[0])
		if id == "" {
			continue
		}
		versionSpec := ""
		if len(parts) >= 2 {
			versionSpec = strings.TrimSpace(parts[1])
		}
		deps = append(deps, Dependency{ID: id, VersionSpec: versionSpec})
	}
	return deps
}

// parseBool converts an OData boolean string ("true"/"false") to bool.
func parseBool(s string) bool {
	return strings.EqualFold(s, "true")
}

// parseTags splits a space-separated tags string into a slice.
// Empty strings and extra whitespace are ignored.
func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

// odataDateFormats lists the datetime formats used by NuGet v2 OData feeds,
// ordered from most specific to least. The community Chocolatey API uses
// fractional seconds without a timezone suffix (e.g. "2026-02-12T18:10:41.9").
var odataDateFormats = []string{
	"2006-01-02T15:04:05.9999999Z", // 7 fractional digits + Z
	"2006-01-02T15:04:05.999999Z",  // 6 fractional digits + Z
	"2006-01-02T15:04:05.9",        // 1-9 fractional digits, no Z (community API)
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	time.RFC3339,
	time.RFC3339Nano,
}

// parseODataDateTime parses an OData DateTime string. Returns zero time on failure.
func parseODataDateTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, format := range odataDateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
