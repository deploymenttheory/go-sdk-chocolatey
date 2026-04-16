package nupkg

import (
	"encoding/xml"
	"fmt"
)

// nuspec XML structs — the .nuspec schema uses
// http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd (and older variants).
// Go's xml package matches elements by local name when no namespace is specified
// on the struct tag, so these work across all nuspec schema versions.

type nuspecXML struct {
	XMLName  xml.Name       `xml:"package"`
	Metadata nuspecMetaXML  `xml:"metadata"`
}

type nuspecMetaXML struct {
	ID               string                `xml:"id"`
	Version          string                `xml:"version"`
	Title            string                `xml:"title"`
	Authors          string                `xml:"authors"`
	Owners           string                `xml:"owners"`
	Description      string                `xml:"description"`
	Summary          string                `xml:"summary"`
	ReleaseNotes     string                `xml:"releaseNotes"`
	Tags             string                `xml:"tags"`
	ProjectURL       string                `xml:"projectUrl"`
	IconURL          string                `xml:"iconUrl"`
	LicenseURL       string                `xml:"licenseUrl"`
	Copyright        string                `xml:"copyright"`
	PackageSourceURL string                `xml:"packageSourceUrl"`
	DocsURL          string                `xml:"docsUrl"`
	BugTrackerURL    string                `xml:"bugTrackerUrl"`
	Dependencies     nuspecDependenciesXML `xml:"dependencies"`
}

type nuspecDependenciesXML struct {
	Items []nuspecDepItemXML `xml:"dependency"`
}

type nuspecDepItemXML struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
}

// parseNuspec parses the raw bytes of a .nuspec XML file.
func parseNuspec(data []byte) (*Nuspec, error) {
	var raw nuspecXML
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("nuspec: %w", err)
	}

	m := raw.Metadata
	deps := make([]NuspecDependency, len(m.Dependencies.Items))
	for i, d := range m.Dependencies.Items {
		deps[i] = NuspecDependency{ID: d.ID, Version: d.Version}
	}

	return &Nuspec{
		ID:               m.ID,
		Version:          m.Version,
		Title:            m.Title,
		Authors:          m.Authors,
		Owners:           m.Owners,
		Description:      m.Description,
		Summary:          m.Summary,
		ReleaseNotes:     m.ReleaseNotes,
		Tags:             m.Tags,
		ProjectURL:       m.ProjectURL,
		IconURL:          m.IconURL,
		LicenseURL:       m.LicenseURL,
		Copyright:        m.Copyright,
		PackageSourceURL: m.PackageSourceURL,
		DocsURL:          m.DocsURL,
		BugTrackerURL:    m.BugTrackerURL,
		Dependencies:     deps,
	}, nil
}
