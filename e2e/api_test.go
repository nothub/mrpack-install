//go:build e2e

package e2e

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nothub/mrpack-install/modrinth/api"
)

// unknownID is a well-formed base62 id that no seeded project or version has.
const unknownID = "zzzzzzzz"

var fabric1218 = versionData{
	VersionNumber: "1.0.0",
	Loaders:       []string{"fabric"},
	GameVersions:  []string{"1.21.8"},
}

func TestGetProject(t *testing.T) {
	t.Parallel()
	const slug = "e2e-api-get-project"
	created := seedMod(t, slug, fabric1218)

	for name, id := range map[string]string{"slug": slug, "id": created.Id} {
		t.Run(name, func(t *testing.T) {
			project, err := api.NewClient(labrinthURL).GetProject(id)
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
		})
	}
}

func TestGetProjectNotFound(t *testing.T) {
	t.Parallel()
	project, err := api.NewClient(labrinthURL).GetProject("e2e-api-missing")
	if err == nil {
		t.Fatalf("GetProject() = %+v, want error", project)
	}
	if !strings.HasPrefix(err.Error(), "http status 404") {
		t.Errorf("error = %q, want http status 404", err)
	}
}

func TestGetProjectsDropsUnknownID(t *testing.T) {
	t.Parallel()
	a := seedMod(t, "e2e-api-get-projects-a", fabric1218)
	b := seedMod(t, "e2e-api-get-projects-b", fabric1218)

	projects, err := api.NewClient(labrinthURL).GetProjects([]string{a.Id, b.Id, unknownID})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, p := range projects {
		ids = append(ids, p.Id)
	}
	if len(ids) != 2 || !slices.Contains(ids, a.Id) || !slices.Contains(ids, b.Id) {
		t.Errorf("project ids = %v, want %q and %q", ids, a.Id, b.Id)
	}
}

func TestCheckProjectValidity(t *testing.T) {
	t.Parallel()
	const slug = "e2e-api-check-validity"
	created := seedMod(t, slug, fabric1218)

	for name, id := range map[string]string{"slug": slug, "id": created.Id} {
		t.Run(name, func(t *testing.T) {
			res, err := api.NewClient(labrinthURL).CheckProjectValidity(id)
			if err != nil {
				t.Fatal(err)
			}
			if res.Id != created.Id {
				t.Errorf("id = %q, want %q", res.Id, created.Id)
			}
		})
	}
}

func TestGetDependencies(t *testing.T) {
	t.Parallel()
	dep := seedMod(t, "e2e-api-dependency", fabric1218)
	v := fabric1218
	v.Dependencies = []dependency{{ProjectID: dep.Id, DependencyType: "required"}}
	const slug = "e2e-api-dependent"
	seedMod(t, slug, v)

	deps, err := api.NewClient(labrinthURL).GetDependencies(slug)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps.Projects) != 1 || deps.Projects[0].Id != dep.Id {
		t.Errorf("dependency projects = %+v, want only %q", deps.Projects, dep.Id)
	}
}

func TestGetProjectVersions(t *testing.T) {
	t.Parallel()
	const slug = "e2e-api-project-versions"
	created := seedMod(t, slug, versionData{
		VersionNumber: "1.0.0",
		Loaders:       []string{"fabric"},
		GameVersions:  []string{"1.20.1"},
	})
	fabric := created.Versions[0]
	// The quilt validator accepts a jar with only fabric.mod.json.
	quilt := uploadVersion(t, created.Id, file{Name: slug + "-quilt.jar", Data: fabricJar(t, slug+"-quilt")}, versionData{
		VersionNumber: "2.0.0",
		Loaders:       []string{"quilt"},
		GameVersions:  []string{"1.21.8"},
	}).Id

	tests := []struct {
		name   string
		params *api.GetProjectVersionsParams
		want   []string
	}{
		{"unfiltered", nil, []string{fabric, quilt}},
		{"loader", &api.GetProjectVersionsParams{Loaders: []string{"quilt"}}, []string{quilt}},
		{"game version", &api.GetProjectVersionsParams{GameVersions: []string{"1.20.1"}}, []string{fabric}},
		{"empty", &api.GetProjectVersionsParams{Loaders: []string{"forge"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := api.NewClient(labrinthURL).GetProjectVersions(slug, tt.params)
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			for _, v := range versions {
				ids = append(ids, v.Id)
			}
			slices.Sort(ids)
			want := slices.Sorted(slices.Values(tt.want))
			if !slices.Equal(ids, want) {
				t.Errorf("version ids = %v, want %v", ids, want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	t.Parallel()
	created := seedMod(t, "e2e-api-get-version", fabric1218)

	version, err := api.NewClient(labrinthURL).GetVersion(created.Versions[0])
	if err != nil {
		t.Fatal(err)
	}
	if version.Id != created.Versions[0] {
		t.Errorf("id = %q, want %q", version.Id, created.Versions[0])
	}
	if version.ProjectId != created.Id {
		t.Errorf("project_id = %q, want %q", version.ProjectId, created.Id)
	}
	if version.VersionNumber != "1.0.0" {
		t.Errorf("version_number = %q, want %q", version.VersionNumber, "1.0.0")
	}
}

func TestGetVersionsDropsUnknownID(t *testing.T) {
	t.Parallel()
	a := seedMod(t, "e2e-api-get-versions-a", fabric1218)
	b := seedMod(t, "e2e-api-get-versions-b", fabric1218)

	versions, err := api.NewClient(labrinthURL).GetVersions([]string{a.Versions[0], b.Versions[0], unknownID})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, v := range versions {
		ids = append(ids, v.Id)
	}
	if len(ids) != 2 || !slices.Contains(ids, a.Versions[0]) || !slices.Contains(ids, b.Versions[0]) {
		t.Errorf("version ids = %v, want %q and %q", ids, a.Versions[0], b.Versions[0])
	}
}

func TestVersionFromHash(t *testing.T) {
	t.Parallel()
	const slug = "e2e-api-version-from-hash"
	jar := fabricJar(t, slug)
	created := createProject(t, api.ModProjectType, slug, file{Name: slug + ".jar", Data: jar}, fabric1218)
	approve(t, created.Id)
	if len(created.Versions) != 1 {
		t.Fatalf("created versions = %v, want one", created.Versions)
	}

	sum := sha1.Sum(jar)
	version, err := api.NewClient(labrinthURL).VersionFromHash(hex.EncodeToString(sum[:]), api.Sha1HashAlgo)
	if err != nil {
		t.Fatal(err)
	}
	if version.Id != created.Versions[0] {
		t.Errorf("id = %q, want %q", version.Id, created.Versions[0])
	}
}

func TestVersionFromMrpackFile(t *testing.T) {
	t.Parallel()
	const slug = "e2e-api-version-from-mrpack"
	pack := mrpack(t, slug)
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)
	if len(created.Versions) != 1 {
		t.Fatalf("created versions = %v, want one", created.Versions)
	}
	path := filepath.Join(t.TempDir(), slug+".mrpack")
	if err := os.WriteFile(path, pack, 0o644); err != nil {
		t.Fatal(err)
	}

	version, err := api.NewClient(labrinthURL).VersionFromMrpackFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if version.Id != created.Versions[0] {
		t.Errorf("id = %q, want %q", version.Id, created.Versions[0])
	}
}
