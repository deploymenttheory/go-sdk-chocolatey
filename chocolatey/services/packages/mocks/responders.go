// Package mocks provides test doubles for the packages service.
package mocks

import (
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/mocks"
)

// PackagesMock is a test double for the packages service.
type PackagesMock struct {
	*mocks.GenericMock
}

// NewPackagesMock returns a PackagesMock backed by a GenericMock.
func NewPackagesMock() *PackagesMock {
	return &PackagesMock{
		GenericMock: mocks.NewXMLMock("PackagesMock"),
	}
}

// RegisterGetByIDLatestMock registers a successful response for GetByID (latest version).
func (m *PackagesMock) RegisterGetByIDLatestMock() {
	m.Register("GET", constants.EndpointPackages, 200, "package_7zip_latest.xml")
}

// RegisterGetByIDAndVersionMock registers a successful response for GetByIDAndVersion.
func (m *PackagesMock) RegisterGetByIDAndVersionMock() {
	m.Register("GET", constants.EndpointPackages, 200, "package_7zip_versioned.xml")
}

// RegisterListVersionsMock registers a successful response for ListVersions.
func (m *PackagesMock) RegisterListVersionsMock() {
	m.Register("GET", constants.EndpointFindPackagesById, 200, "packages_versions.xml")
}

// RegisterSearchMock registers a successful response for Search.
func (m *PackagesMock) RegisterSearchMock() {
	m.Register("GET", constants.EndpointSearch, 200, "packages_search.xml")
}

// RegisterNotFoundMock registers an empty feed response (package not found).
func (m *PackagesMock) RegisterNotFoundMock() {
	m.Register("GET", constants.EndpointPackages, 200, "packages_empty.xml")
}

// RegisterServerErrorMock registers a 500 Internal Server Error for /Packages().
func (m *PackagesMock) RegisterServerErrorMock() {
	m.RegisterError("GET", constants.EndpointPackages, 500, "internal server error")
}

// RegisterSearchServerErrorMock registers a 500 Internal Server Error for /Search().
func (m *PackagesMock) RegisterSearchServerErrorMock() {
	m.RegisterError("GET", constants.EndpointSearch, 500, "internal server error")
}
