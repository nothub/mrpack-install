package modrinth

import (
	"testing"
)

func Test_GetProject_Success(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	project, err := c.GetProject("fabric-api")
	if err != nil {
		t.Fatal(err)
	}
	if project.Slug != "fabric-api" {
		t.Fatal("wrong slug!")
	}
	if project.ProjectType != "mod" {
		t.Fatal("wrong type!")
	}
}

func Test_GetProject_404(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	_, err := c.GetProject("x")
	if err.Error() != "http status 404" {
		t.Fatal("wrong status!")
	}
}

func TestClient_GetProjects_Count(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	projects, err := c.GetProjects([]string{"P7dR8mSH", "XxWD5pD3", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatal("wrong count!")
	}
}

func TestClient_GetProjects_Slug(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	projects, err := c.GetProjects([]string{"P7dR8mSH"})
	if err != nil {
		t.Fatal(err)
	}
	if projects[0].Slug != "fabric-api" {
		t.Fatal("wrong slug!")
	}
}

func TestClient_CheckProjectValidity_Slug(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	response, err := c.CheckProjectValidity("fabric-api")
	if err != nil {
		t.Fatal(err)
	}
	if response.Id != "P7dR8mSH" {
		t.Fatal("wrong id!")
	}
}

func TestClient_CheckProjectValidity_Id(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	response, err := c.CheckProjectValidity("P7dR8mSH")
	if err != nil {
		t.Fatal(err)
	}
	if response.Id != "P7dR8mSH" {
		t.Fatal("wrong id!")
	}
}

func TestClient_GetDependencies(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	dependencies, err := c.GetDependencies("rinthereout")
	if err != nil {
		t.Fatal(err)
	}
	if len(dependencies.Projects) < 1 {
		t.Fatal("wrong count!")
	}
}

func TestClient_GetProjectVersions_Count(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	versions, err := c.GetProjectVersions("fabric-api", &GetProjectVersionsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) < 1 {
		t.Fatal("wrong count!")
	}
}

func TestClient_GetProjectVersions_Filter_Results(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	versions, err := c.GetProjectVersions("fabric-api", &GetProjectVersionsParams{
		GameVersions: []string{"1.16.5"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) < 1 {
		t.Fatal("wrong count!")
	}
}

func TestClient_GetProjectVersions_Filter_NoResults(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	versions, err := c.GetProjectVersions("fabric-api", &GetProjectVersionsParams{
		Loaders: []string{"forge"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) > 0 {
		t.Fatal("wrong count!")
	}
}

func TestClient_GetVersion(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	version, err := c.GetVersion("IQ3UGSc2")
	if err != nil {
		t.Fatal(err)
	}
	if version.ProjectId != "P7dR8mSH" {
		t.Fatal("wrong parent id!")
	}
}

func TestClient_GetVersions(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	versions, err := c.GetVersions([]string{"IQ3UGSc2", "DrzwF8io", "foobar"})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatal("wrong count!")
	}
}

func TestClient_VersionFromHash(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	version, err := c.VersionFromHash("619e250c133106bacc3e3b560839bd4b324dfda8", "sha1")
	if err != nil {
		t.Fatal(err)
	}
	if version.Id != "d5nXweHE" {
		t.Fatal("wrong id!")
	}
}

func TestClient_GetLatestGameVersion(t *testing.T) {
	t.Parallel()
	c := NewClient("api.modrinth.com")
	version, err := c.GetLatestGameVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version == "" {
		t.Fatal("result missing!")
	}
}
