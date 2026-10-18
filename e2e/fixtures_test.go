//go:build e2e

package e2e

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/nothub/mrpack-install/modrinth/api"
)

// file is an upload to labrinth. The extension of Name decides how labrinth
// validates it, and labrinth rejects a file whose SHA-1 another project
// already has.
type file struct {
	Name string
	Data []byte
}

// versionData is the caller's part of labrinth's v2 version create data.
type versionData struct {
	VersionNumber string       `json:"version_number"`
	Loaders       []string     `json:"loaders"`
	GameVersions  []string     `json:"game_versions"`
	Dependencies  []dependency `json:"dependencies"`
}

type dependency struct {
	ProjectID      string `json:"project_id,omitempty"`
	VersionID      string `json:"version_id,omitempty"`
	DependencyType string `json:"dependency_type"`
}

// versionCreate is InitialVersionData in labrinth's routes/v2/version_creation.rs.
type versionCreate struct {
	versionData
	ProjectID      string   `json:"project_id,omitempty"`
	FileParts      []string `json:"file_parts"`
	VersionTitle   string   `json:"version_title"`
	ReleaseChannel string   `json:"release_channel"`
	Featured       bool     `json:"featured"`
}

func newVersionCreate(f file, v versionData) versionCreate {
	// labrinth has no default for dependencies, so null fails to parse.
	if v.Dependencies == nil {
		v.Dependencies = []dependency{}
	}
	return versionCreate{
		versionData:    v,
		FileParts:      []string{f.Name},
		VersionTitle:   v.VersionNumber,
		ReleaseChannel: "release",
	}
}

// createProject creates a project owned by the seed admin, with one initial
// version holding f. The project stays in review until approve.
func createProject(t testing.TB, projectType api.ProjectType, slug string, f file, v versionData) *api.Project {
	t.Helper()
	data := struct {
		Title           string          `json:"title"`
		ProjectType     api.ProjectType `json:"project_type"`
		Slug            string          `json:"slug"`
		Description     string          `json:"description"`
		Body            string          `json:"body"`
		ClientSide      string          `json:"client_side"`
		ServerSide      string          `json:"server_side"`
		InitialVersions []versionCreate `json:"initial_versions"`
		Categories      []string        `json:"categories"`
		LicenseID       string          `json:"license_id"`
		IsDraft         bool            `json:"is_draft"`
	}{
		Title:           slug,
		ProjectType:     projectType,
		Slug:            slug,
		Description:     "e2e fixture",
		ClientSide:      "required",
		ServerSide:      "required",
		InitialVersions: []versionCreate{newVersionCreate(f, v)},
		Categories:      []string{},
		LicenseID:       "MIT",
	}
	var project api.Project
	upload(t, "/v2/project", data, f, &project)
	return &project
}

// uploadVersion adds a version holding f to an existing project.
func uploadVersion(t testing.TB, projectID string, f file, v versionData) *api.Version {
	t.Helper()
	data := newVersionCreate(f, v)
	data.ProjectID = projectID
	var version api.Version
	upload(t, "/v2/version", data, f, &version)
	return &version
}

// approve makes a project visible to unauthenticated reads.
func approve(t testing.TB, projectID string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, labrinthURL+"/v2/project/"+projectID,
		strings.NewReader(`{"status":"approved"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", adminPAT)
	req.Header.Set("Content-Type", "application/json")
	do(t, req, http.StatusNoContent, nil)
}

// upload posts data and f as the multipart form labrinth's create routes
// take, with data as the first part.
func upload(t testing.TB, path string, data any, f file, out any) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	dw, err := mw.CreateFormField("data")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(dw).Encode(data); err != nil {
		t.Fatal(err)
	}
	fw, err := mw.CreateFormFile(f.Name, f.Name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(f.Data); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, labrinthURL+path, &body)
	if err != nil {
		t.Fatal(err)
	}
	// A regular user may create 10 projects a day, which the suite
	// exceeds on a fresh database. Admins have no limits.
	req.Header.Set("Authorization", adminPAT)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	do(t, req, http.StatusOK, out)
}

// do sends req and fails t unless labrinth answers with status want. A
// non-nil out receives the decoded JSON body.
func do(t testing.TB, req *http.Request, want int, out any) {
	t.Helper()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != want {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("%s %s: status = %d, want %d: %s", req.Method, req.URL.Path, res.StatusCode, want, b)
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatalf("%s %s: decode: %v", req.Method, req.URL.Path, err)
		}
	}
}

// fabricJar returns a minimal Fabric mod jar. Its bytes depend on modID, so
// jars with different ids have different hashes.
func fabricJar(t testing.TB, modID string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("fabric.mod.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(w, `{"schemaVersion":1,"id":%q,"version":"1.0.0"}`, modID)
	if err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
