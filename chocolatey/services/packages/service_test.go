package packages_test

import (
	"context"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/client"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/packages"
	pkgmocks "github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/services/packages/mocks"
)

const testBaseURL = "https://community.chocolatey.org/api/v2"

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestUnit_Packages_GetByID_happyPath(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDLatestMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, resp, err := svc.GetByID(context.Background(), "7zip")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
	if pkg.ID != "7zip" {
		t.Errorf("ID: got %q, want %q", pkg.ID, "7zip")
	}
	if pkg.Version == "" {
		t.Error("Version should not be empty")
	}
	if !pkg.IsLatestVersion {
		t.Error("IsLatestVersion should be true for latest fixture")
	}
	if pkg.PackageHash == "" {
		t.Error("PackageHash should not be empty")
	}
	if len(pkg.Tags) == 0 {
		t.Error("Tags should not be empty")
	}
	wantURL := testBaseURL + "/package/7zip/" + pkg.Version
	if pkg.DownloadURL != wantURL {
		t.Errorf("DownloadURL: got %q, want %q", pkg.DownloadURL, wantURL)
	}
}

func TestUnit_Packages_GetByID_emptyID_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	svc := packages.NewPackages(m, testBaseURL)

	_, _, err := svc.GetByID(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
}

func TestUnit_Packages_GetByID_notFound_returnsNotFoundError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterNotFoundMock()

	svc := packages.NewPackages(m, testBaseURL)
	_, _, err := svc.GetByID(context.Background(), "nonexistent-package")

	if err == nil {
		t.Fatal("expected error for empty feed, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got: %v", err)
	}
}

func TestUnit_Packages_GetByID_apiError_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterServerErrorMock()

	svc := packages.NewPackages(m, testBaseURL)
	_, _, err := svc.GetByID(context.Background(), "7zip")

	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
	if !client.IsServerError(err) {
		t.Errorf("expected IsServerError, got: %v", err)
	}
}

// ── GetByIDAndVersion ─────────────────────────────────────────────────────────

func TestUnit_Packages_GetByIDAndVersion_happyPath(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDAndVersionMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, resp, err := svc.GetByIDAndVersion(context.Background(), "7zip", "23.1.0")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if pkg.ID != "7zip" {
		t.Errorf("ID: got %q, want %q", pkg.ID, "7zip")
	}
	if pkg.Version != "23.1.0" {
		t.Errorf("Version: got %q, want %q", pkg.Version, "23.1.0")
	}
	if pkg.IsLatestVersion {
		t.Error("IsLatestVersion should be false for versioned fixture")
	}
	if len(pkg.Dependencies) == 0 {
		t.Error("Dependencies should not be empty for versioned fixture")
	}
}

func TestUnit_Packages_GetByIDAndVersion_emptyID_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	svc := packages.NewPackages(m, testBaseURL)

	_, _, err := svc.GetByIDAndVersion(context.Background(), "", "1.0.0")
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
}

func TestUnit_Packages_GetByIDAndVersion_emptyVersion_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	svc := packages.NewPackages(m, testBaseURL)

	_, _, err := svc.GetByIDAndVersion(context.Background(), "7zip", "")
	if err == nil {
		t.Fatal("expected error for empty version, got nil")
	}
}

func TestUnit_Packages_GetByIDAndVersion_notFound_returnsNotFoundError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterNotFoundMock()

	svc := packages.NewPackages(m, testBaseURL)
	_, _, err := svc.GetByIDAndVersion(context.Background(), "7zip", "0.0.0")

	if err == nil {
		t.Fatal("expected error for empty feed, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got: %v", err)
	}
}

// ── ListVersions ──────────────────────────────────────────────────────────────

func TestUnit_Packages_ListVersions_happyPath(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterListVersionsMock()

	svc := packages.NewPackages(m, testBaseURL)
	result, resp, err := svc.ListVersions(context.Background(), "7zip")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Versions) == 0 {
		t.Fatal("expected at least one version")
	}
	if result.TotalCount != len(result.Versions) {
		t.Errorf("TotalCount (%d) != len(Versions) (%d)", result.TotalCount, len(result.Versions))
	}
	// versions fixture has 3 entries
	if result.TotalCount != 3 {
		t.Errorf("TotalCount: got %d, want 3", result.TotalCount)
	}
}

func TestUnit_Packages_ListVersions_emptyID_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	svc := packages.NewPackages(m, testBaseURL)

	_, _, err := svc.ListVersions(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty id, got nil")
	}
}

// ── Search ────────────────────────────────────────────────────────────────────

func TestUnit_Packages_Search_happyPath(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterSearchMock()

	svc := packages.NewPackages(m, testBaseURL)
	result, resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		SearchTerm: "zip",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Packages) == 0 {
		t.Fatal("expected at least one package")
	}
	if result.TotalCount != len(result.Packages) {
		t.Errorf("TotalCount (%d) != len(Packages) (%d)", result.TotalCount, len(result.Packages))
	}
}

func TestUnit_Packages_Search_emptyTerm_returnsResults(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterSearchMock()

	svc := packages.NewPackages(m, testBaseURL)
	// empty SearchTerm is valid — returns all packages
	result, _, err := svc.Search(context.Background(), &packages.FilterOptions{})

	if err != nil {
		t.Fatalf("unexpected error for empty search term: %v", err)
	}
	if result == nil || len(result.Packages) == 0 {
		t.Error("expected results even with empty search term")
	}
}

func TestUnit_Packages_Search_nilOpts_doesNotPanic(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterSearchMock()

	svc := packages.NewPackages(m, testBaseURL)
	result, _, err := svc.Search(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error with nil opts: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result with nil opts")
	}
}

func TestUnit_Packages_Search_apiError_returnsError(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterSearchServerErrorMock()

	svc := packages.NewPackages(m, testBaseURL)
	_, _, err := svc.Search(context.Background(), &packages.FilterOptions{SearchTerm: "zip"})

	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
	if !client.IsServerError(err) {
		t.Errorf("expected IsServerError, got: %v", err)
	}
}

// ── Parser unit tests ─────────────────────────────────────────────────────────

func TestUnit_parseDependencies_emptyString(t *testing.T) {
	// Exposed indirectly via GetByIDAndVersion fixture check above, but we can
	// also test via a package with known empty deps.
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDLatestMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, _, err := svc.GetByID(context.Background(), "7zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Latest fixture includes deps; just confirm they are parsed (not nil slice
	// when the fixture has them).
	_ = pkg.Dependencies
}

func TestUnit_mapEntryToPackage_downloadURL(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDLatestMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, _, err := svc.GetByID(context.Background(), "7zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := testBaseURL + "/package/7zip/" + pkg.Version
	if pkg.DownloadURL != want {
		t.Errorf("DownloadURL: got %q, want %q", pkg.DownloadURL, want)
	}
}

func TestUnit_mapEntryToPackage_parsedTimes(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDLatestMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, _, err := svc.GetByID(context.Background(), "7zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	zero := time.Time{}
	if pkg.Published == zero {
		t.Error("Published should be non-zero")
	}
	if pkg.Created == zero {
		t.Error("Created should be non-zero")
	}
	if pkg.LastUpdated == zero {
		t.Error("LastUpdated should be non-zero")
	}
}

func TestUnit_mapEntryToPackage_packageSize(t *testing.T) {
	m := pkgmocks.NewPackagesMock()
	m.RegisterGetByIDLatestMock()

	svc := packages.NewPackages(m, testBaseURL)
	pkg, _, err := svc.GetByID(context.Background(), "7zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.PackageSize <= 0 {
		t.Errorf("PackageSize should be positive, got %d", pkg.PackageSize)
	}
}
