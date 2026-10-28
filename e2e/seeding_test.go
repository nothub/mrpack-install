package e2e

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/nothub/mrpack-install/modrinth/api"
)

func TestSeededModServesFile(t *testing.T) {
	t.Parallel()
	const slug = "e2e-seeded-mod"
	jar := fabricJar(t, slug)
	created := createProject(t, api.ModProjectType, slug, file{Name: slug + ".jar", Data: jar}, versionData{
		VersionNumber: "1.0.0",
		Loaders:       []string{"fabric"},
		GameVersions:  []string{"1.21.8"},
	})
	approve(t, created.Id)

	client := api.NewClient(labrinthURL)
	project, err := client.GetProject(slug)
	if err != nil {
		t.Fatal(err)
	}
	if project.Id != created.Id {
		t.Errorf("id = %q, want %q", project.Id, created.Id)
	}
	if project.Slug != slug {
		t.Errorf("slug = %q, want %q", project.Slug, slug)
	}
	if project.ProjectType != api.ModProjectType {
		t.Errorf("project_type = %q, want %q", project.ProjectType, api.ModProjectType)
	}
	if len(project.Versions) != 1 {
		t.Fatalf("versions = %v, want one", project.Versions)
	}

	version, err := client.GetVersion(project.Versions[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(version.Files) != 1 {
		t.Fatalf("files = %v, want one", version.Files)
	}
	res, err := http.Get(version.Files[0].Url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want %d", version.Files[0].Url, res.StatusCode, http.StatusOK)
	}
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, jar) {
		t.Errorf("cdn served %d bytes that differ from the %d uploaded", len(got), len(jar))
	}
	sum := sha512.Sum512(jar)
	if want := hex.EncodeToString(sum[:]); version.Files[0].Hashes.Sha512 != want {
		t.Errorf("sha512 = %q, want %q", version.Files[0].Hashes.Sha512, want)
	}
}

func TestUploadedVersionIsListed(t *testing.T) {
	t.Parallel()
	const slug = "e2e-uploaded-version"
	v := versionData{
		VersionNumber: "1.0.0",
		Loaders:       []string{"fabric"},
		GameVersions:  []string{"1.21.8"},
	}
	created := createProject(t, api.ModProjectType, slug, file{Name: slug + ".jar", Data: fabricJar(t, slug)}, v)
	if len(created.Versions) != 1 {
		t.Fatalf("created versions = %v, want one", created.Versions)
	}
	approve(t, created.Id)

	v.VersionNumber = "2.0.0"
	uploaded := uploadVersion(t, created.Id, file{Name: slug + "-2.jar", Data: fabricJar(t, slug+"-2")}, v)

	versions, err := api.NewClient(labrinthURL).GetProjectVersions(slug, nil)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, version := range versions {
		ids = append(ids, version.Id)
	}
	if len(ids) != 2 || !slices.Contains(ids, created.Versions[0]) || !slices.Contains(ids, uploaded.Id) {
		t.Errorf("version ids = %v, want %q and %q", ids, created.Versions[0], uploaded.Id)
	}
}
