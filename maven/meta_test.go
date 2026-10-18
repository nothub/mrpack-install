package maven

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// testdata/quilt-installer/maven-metadata.xml is a verbatim copy of
// https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/maven-metadata.xml
func TestFetchMetadata(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	t.Cleanup(srv.Close)

	m, err := FetchMetadata(srv.URL + "/quilt-installer/maven-metadata.xml")
	if err != nil {
		t.Fatal(err)
	}
	if m.ArtifactId != "quilt-installer" {
		t.Errorf("ArtifactId = %q, want %q", m.ArtifactId, "quilt-installer")
	}
	if m.Versioning.Latest != "0.15.1" {
		t.Errorf("Versioning.Latest = %q, want %q", m.Versioning.Latest, "0.15.1")
	}
	if m.Versioning.Release != "0.15.1" {
		t.Errorf("Versioning.Release = %q, want %q", m.Versioning.Release, "0.15.1")
	}
	if len(m.Versioning.Versions) == 0 {
		t.Error("Versioning.Versions is empty")
	}
}
