package update

import (
	"archive/zip"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/packstate"
)

func hashes(content string) modrinth.Hashes {
	s1 := sha1.Sum([]byte(content))
	s512 := sha512.Sum512([]byte(content))
	return modrinth.Hashes{Sha1: hex.EncodeToString(s1[:]), Sha512: hex.EncodeToString(s512[:])}
}

func sha1Only(content string) modrinth.Hashes {
	return modrinth.Hashes{Sha1: hashes(content).Sha1}
}

// writePack stores index and the given entries in an archive and returns
// the index and the archive path.
func writePack(t *testing.T, index mrpack.Index, entries map[string]string) (*mrpack.Index, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "pack.mrpack")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	iw, err := w.Create("modrinth.index.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(iw).Encode(index); err != nil {
		t.Fatal(err)
	}
	for name, data := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return &index, path
}

// serverDir returns a temp dir holding files, by path.
func serverDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, data := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func str(s string) *string { return &s }

func ptr(h modrinth.Hashes) *modrinth.Hashes { return &h }

func state(records map[string]modrinth.Hashes) *packstate.Schema {
	return &packstate.Schema{SchemaVersion: 1, Hashes: records}
}

func indexFile(path string, server modrinth.EnvSupport, content string) mrpack.File {
	return mrpack.File{
		Path:      path,
		Hashes:    hashes(content),
		Env:       mrpack.Env{Client: modrinth.RequiredEnvSupport, Server: server},
		Downloads: []string{"https://cdn.example.org/" + path},
	}
}

// TestComputeBehaviorTable has one row per cell of the replacement rules.
// The new pack ships x as an override when shipped is set, the old state
// records x when old is set, and the server dir holds x when disk is set.
func TestComputeBehaviorTable(t *testing.T) {
	tests := []struct {
		name    string
		shipped *string
		old     *modrinth.Hashes
		disk    *string
		want    []Action
	}{
		{"ships, disk equals new", str("new"), nil, str("new"),
			[]Action{{Path: "x", Op: OpNone}}},
		{"ships, disk missing", str("new"), nil, nil,
			[]Action{{Path: "x", Op: OpWrite}}},
		{"ships, recorded, disk missing", str("new"), ptr(hashes("old")), nil,
			[]Action{{Path: "x", Op: OpWrite, Recorded: true}}},
		{"ships, disk equals old record", str("new"), ptr(hashes("old")), str("old"),
			[]Action{{Path: "x", Op: OpWrite, Recorded: true}}},
		{"ships, disk equals old SHA-1 record", str("new"), ptr(sha1Only("old")), str("old"),
			[]Action{{Path: "x", Op: OpWrite, Recorded: true}}},
		{"ships, disk drifted from old record", str("new"), ptr(hashes("old")), str("edited"),
			[]Action{{Path: "x", Op: OpBackupWrite, Recorded: true}}},
		{"ships, disk drifted from old SHA-1 record", str("new"), ptr(sha1Only("old")), str("edited"),
			[]Action{{Path: "x", Op: OpBackupWrite, Recorded: true}}},
		{"ships, disk file without old record", str("new"), nil, str("mine"),
			[]Action{{Path: "x", Op: OpBackupWrite}}},
		{"dropped, disk missing", nil, ptr(hashes("old")), nil,
			[]Action{{Path: "x", Op: OpNone, Recorded: true}}},
		{"dropped, disk equals old record", nil, ptr(hashes("old")), str("old"),
			[]Action{{Path: "x", Op: OpDelete, Recorded: true}}},
		{"dropped, disk equals old SHA-1 record", nil, ptr(sha1Only("old")), str("old"),
			[]Action{{Path: "x", Op: OpDelete, Recorded: true}}},
		{"dropped, disk drifted from old record", nil, ptr(hashes("old")), str("edited"),
			[]Action{{Path: "x", Op: OpBackupDelete, Recorded: true}}},
		{"dropped, disk drifted from old SHA-1 record", nil, ptr(sha1Only("old")), str("edited"),
			[]Action{{Path: "x", Op: OpBackupDelete, Recorded: true}}},
		{"never shipped nor recorded", nil, nil, str("mine"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entries := map[string]string{"overrides/keep": "keep"}
			if tt.shipped != nil {
				entries["overrides/x"] = *tt.shipped
			}
			index, archive := writePack(t, mrpack.Index{Name: "pack"}, entries)
			records := map[string]modrinth.Hashes{"keep": hashes("keep")}
			if tt.old != nil {
				records["x"] = *tt.old
			}
			disk := map[string]string{"keep": "keep"}
			if tt.disk != nil {
				disk["x"] = *tt.disk
			}

			plan, err := Compute(serverDir(t, disk), state(records), index, archive, packstate.Options{})
			if err != nil {
				t.Fatal(err)
			}

			want := append([]Action{{Path: "keep", Op: OpNone, Recorded: true}}, tt.want...)
			if !reflect.DeepEqual(plan.Actions, want) {
				t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
			}
		})
	}
}

func TestComputeSelectsIndexFiles(t *testing.T) {
	t.Parallel()
	index, archive := writePack(t, mrpack.Index{Name: "pack", Files: []mrpack.File{
		indexFile("mods/required.jar", modrinth.RequiredEnvSupport, "required"),
		indexFile("mods/opt-a.jar", modrinth.OptionalEnvSupport, "opt-a"),
		indexFile("mods/opt-b.jar", modrinth.OptionalEnvSupport, "opt-b"),
		indexFile("mods/client.jar", modrinth.UnsupportedEnvSupport, "client"),
	}}, nil)
	// opt-b was installed by the previous run and is deselected now.
	old := state(map[string]modrinth.Hashes{"mods/opt-b.jar": hashes("opt-b")})
	dir := serverDir(t, map[string]string{"mods/opt-b.jar": "opt-b"})

	plan, err := Compute(dir, old, index, archive, packstate.Options{OptionalSelect: []string{`opt-a`}})
	if err != nil {
		t.Fatal(err)
	}

	want := []Action{
		{Path: "mods/opt-a.jar", Op: OpWrite},
		{Path: "mods/opt-b.jar", Op: OpDelete, Recorded: true},
		{Path: "mods/required.jar", Op: OpWrite},
	}
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
	}
}

func TestComputeOverrideWinsOverIndexFile(t *testing.T) {
	t.Parallel()
	index, archive := writePack(t, mrpack.Index{Name: "pack", Files: []mrpack.File{
		indexFile("config/x.txt", modrinth.RequiredEnvSupport, "index"),
	}}, map[string]string{"server-overrides/config/x.txt": "override"})
	dir := serverDir(t, map[string]string{"config/x.txt": "override"})

	plan, err := Compute(dir, nil, index, archive, packstate.Options{})
	if err != nil {
		t.Fatal(err)
	}

	want := []Action{{Path: "config/x.txt", Op: OpNone}}
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
	}
}

// A pack that renames mods/Foo.jar to mods/foo.jar must not delete the new
// file on a case-insensitive filesystem, where both names are one file.
func TestComputeCaseOnlyRename(t *testing.T) {
	t.Parallel()
	index, archive := writePack(t, mrpack.Index{Name: "pack"}, map[string]string{"overrides/mods/foo.jar": "new"})
	old := state(map[string]modrinth.Hashes{"mods/Foo.jar": hashes("old")})
	dir := serverDir(t, map[string]string{"mods/Foo.jar": "old"})
	_, err := os.Stat(filepath.Join(dir, "mods", "foo.jar"))
	caseInsensitive := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}

	plan, err := Compute(dir, old, index, archive, packstate.Options{})
	if err != nil {
		t.Fatal(err)
	}

	want := []Action{
		{Path: "mods/Foo.jar", Op: OpDelete, Recorded: true},
		{Path: "mods/foo.jar", Op: OpWrite},
	}
	if caseInsensitive {
		t.Log("case-insensitive filesystem")
		want = []Action{{Path: "mods/foo.jar", Op: OpWrite}}
	}
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
	}
}

func TestComputeUppercaseHexMatches(t *testing.T) {
	t.Parallel()
	upper := hashes("new")
	upper.Sha512 = strings.ToUpper(upper.Sha512)
	file := indexFile("mods/a.jar", modrinth.RequiredEnvSupport, "new")
	file.Hashes = upper
	index, archive := writePack(t, mrpack.Index{Name: "pack", Files: []mrpack.File{file}}, nil)
	dir := serverDir(t, map[string]string{"mods/a.jar": "new"})

	plan, err := Compute(dir, nil, index, archive, packstate.Options{})
	if err != nil {
		t.Fatal(err)
	}

	want := []Action{{Path: "mods/a.jar", Op: OpNone}}
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
	}
}

func TestComputeDirectoryAtPackPathFails(t *testing.T) {
	t.Parallel()
	index, archive := writePack(t, mrpack.Index{Name: "pack"}, map[string]string{"overrides/config": "file"})
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0755); err != nil {
		t.Fatal(err)
	}

	plan, err := Compute(dir, nil, index, archive, packstate.Options{})

	if err == nil {
		t.Errorf("Compute() = %+v, want an error", plan)
	}
}

func TestComputeUnsafePathFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		old   *packstate.Schema
		files []mrpack.File
	}{
		{"old state", state(map[string]modrinth.Hashes{"../evil.jar": hashes("old")}), nil},
		{"new pack", nil, []mrpack.File{indexFile("../evil.jar", modrinth.RequiredEnvSupport, "new")}},
		{"staging dir", nil, []mrpack.File{indexFile(".mrpack-install-staging/x.jar", modrinth.RequiredEnvSupport, "new")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			index, archive := writePack(t, mrpack.Index{Name: "pack", Files: tt.files}, nil)

			plan, err := Compute(serverDir(t, nil), tt.old, index, archive, packstate.Options{})

			if err == nil {
				t.Errorf("Compute() = %+v, want an error", plan)
			}
		})
	}
}

func TestComputeInstallServer(t *testing.T) {
	t.Parallel()
	deps := mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.17.2"}
	tests := []struct {
		name string
		old  *packstate.Schema
		want bool
	}{
		{"no old state", nil, true},
		{"same dependencies", &packstate.Schema{SchemaVersion: 1, Deps: deps}, false},
		{"loader changed", &packstate.Schema{SchemaVersion: 1, Deps: mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.16.14"}}, true},
		{"minecraft changed", &packstate.Schema{SchemaVersion: 1, Deps: mrpack.Deps{Minecraft: "1.21.7", Fabric: "0.17.2"}}, true},
		{"loader switched", &packstate.Schema{SchemaVersion: 1, Deps: mrpack.Deps{Minecraft: "1.21.8", Quilt: "0.29.1"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			index, archive := writePack(t, mrpack.Index{Name: "pack", Deps: deps}, nil)

			plan, err := Compute(serverDir(t, nil), tt.old, index, archive, packstate.Options{})
			if err != nil {
				t.Fatal(err)
			}

			if plan.InstallServer != tt.want {
				t.Errorf("InstallServer = %v, want %v", plan.InstallServer, tt.want)
			}
		})
	}
}

func TestComputeSkipsExcludedPaths(t *testing.T) {
	t.Parallel()
	index, archive := writePack(t, mrpack.Index{Name: "pack", Files: []mrpack.File{
		indexFile("mods/tuned.jar", modrinth.RequiredEnvSupport, "new"),
		indexFile("mods/other.jar", modrinth.RequiredEnvSupport, "new"),
	}}, map[string]string{"overrides/config/x.txt": "new"})
	old := state(map[string]modrinth.Hashes{
		"mods/tuned.jar":  hashes("old"),
		"config/x.txt":    hashes("old"),
		"config/gone.txt": hashes("old"),
	})
	dir := serverDir(t, map[string]string{"mods/tuned.jar": "edited", "config/x.txt": "old", "config/gone.txt": "old"})
	opts := packstate.Options{Exclude: []string{"config/**", "mods/tuned.jar"}}

	plan, err := Compute(dir, old, index, archive, opts)
	if err != nil {
		t.Fatal(err)
	}

	want := []Action{{Path: "mods/other.jar", Op: OpWrite}}
	if !reflect.DeepEqual(plan.Actions, want) {
		t.Errorf("Actions = %+v, want %+v", plan.Actions, want)
	}
}
