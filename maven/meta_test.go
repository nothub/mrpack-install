package maven

import (
	"testing"
)

func TestFetchMetadata(t *testing.T) {
	m, err := FetchMetadata("https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/maven-metadata.xml")
	if err != nil {
		t.Fatal(err.Error())
	}
	if m.ArtifactId != "quilt-installer" {
		t.Fatal("wrong artifact id")
	}
	t.Logf("%++v", m)
}
