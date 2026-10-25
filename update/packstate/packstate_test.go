package packstate

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
)

// hashes of the bytes "server"
const serverSha1 = "3de4f901fffb30ac720b0e7eb654b4faa2dd03fa"
const serverSha512 = "fb7daafb37d21221f75d0451fa35dfaacfca509150cecad0e40217b593c5b47566e80c1dd1f74556b3c46a357b699c860976372361a99332934e720a586b7786"

func indexFile(path string, server modrinth.EnvSupport, hash string) mrpack.File {
	return mrpack.File{
		Path:      path,
		Hashes:    modrinth.Hashes{Sha1: hash + "-sha1", Sha512: hash + "-sha512"},
		Env:       mrpack.Env{Client: modrinth.RequiredEnvSupport, Server: server},
		Downloads: []string{"https://cdn.example.org/" + path},
	}
}

// writePack stores index and the given entries in an archive and returns
// its path.
func writePack(t *testing.T, index mrpack.Index, entries [][2]string) string {
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
	for _, e := range entries {
		fw, err := w.Create(e[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(e[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func archiveSha512(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha512.Sum512(b)
	return hex.EncodeToString(sum[:])
}

func noIds(string) (Ids, error) {
	return Ids{}, errors.New("not on modrinth")
}

func TestFromArchiveRecordsFilesInPlace(t *testing.T) {
	t.Parallel()
	pack := writePack(t, mrpack.Index{
		Format:  1,
		Game:    mrpack.Minecraft,
		Version: "1.2.0",
		Name:    "Example",
		Files: []mrpack.File{
			indexFile("mods/required.jar", modrinth.RequiredEnvSupport, "required"),
			indexFile("mods/opt-a.jar", modrinth.OptionalEnvSupport, "opt-a"),
			indexFile("mods/opt-b.jar", modrinth.OptionalEnvSupport, "opt-b"),
			indexFile("mods/client.jar", modrinth.UnsupportedEnvSupport, "client"),
			indexFile("config/shared.txt", modrinth.RequiredEnvSupport, "shadowed"),
		},
		Deps: mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.17.2"},
	}, [][2]string{
		{"server-overrides/config/shared.txt", "server"},
		{"overrides/config/extra.txt", "server"},
		{"client-overrides/config/client.txt", "client"},
	})
	opts := Options{OptionalSelect: []string{`opt-a`}}

	state, err := FromArchive(pack, opts, noIds)
	if err != nil {
		t.Fatal(err)
	}

	if state.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", state.SchemaVersion)
	}
	if state.Name != "Example" || state.Version != "1.2.0" {
		t.Errorf("Name, Version = %q, %q, want %q, %q", state.Name, state.Version, "Example", "1.2.0")
	}
	if want := archiveSha512(t, pack); state.ArchiveSha512 != want {
		t.Errorf("ArchiveSha512 = %q, want %q", state.ArchiveSha512, want)
	}
	if !reflect.DeepEqual(state.Options, opts) {
		t.Errorf("Options = %+v, want %+v", state.Options, opts)
	}
	if state.Deps != (mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.17.2"}) {
		t.Errorf("Deps = %+v", state.Deps)
	}
	server := modrinth.Hashes{Sha1: serverSha1, Sha512: serverSha512}
	want := map[string]modrinth.Hashes{
		"mods/required.jar": {Sha1: "required-sha1", Sha512: "required-sha512"},
		"mods/opt-a.jar":    {Sha1: "opt-a-sha1", Sha512: "opt-a-sha512"},
		"config/shared.txt": server,
		"config/extra.txt":  server,
	}
	if !reflect.DeepEqual(state.Hashes, want) {
		t.Errorf("Hashes = %v, want %v", state.Hashes, want)
	}
}

func TestFromArchiveOptionalDisableAll(t *testing.T) {
	t.Parallel()
	pack := writePack(t, mrpack.Index{
		Name: "Example",
		Files: []mrpack.File{
			indexFile("mods/required.jar", modrinth.RequiredEnvSupport, "required"),
			indexFile("mods/opt-a.jar", modrinth.OptionalEnvSupport, "opt-a"),
		},
	}, nil)

	state, err := FromArchive(pack, Options{OptionalDisableAll: true}, noIds)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]modrinth.Hashes{
		"mods/required.jar": {Sha1: "required-sha1", Sha512: "required-sha512"},
	}
	if !reflect.DeepEqual(state.Hashes, want) {
		t.Errorf("Hashes = %v, want %v", state.Hashes, want)
	}
}

func TestFromArchiveIds(t *testing.T) {
	t.Parallel()
	pack := writePack(t, mrpack.Index{Name: "Example"}, nil)
	var asked string
	lookup := func(sha512 string) (Ids, error) {
		asked = sha512
		return Ids{Slug: "example", ProjectId: "P1", VersionId: "V1"}, nil
	}

	state, err := FromArchive(pack, Options{}, lookup)
	if err != nil {
		t.Fatal(err)
	}

	if want := archiveSha512(t, pack); asked != want {
		t.Errorf("lookup got %q, want the archive SHA-512 %q", asked, want)
	}
	if state.Slug != "example" || state.ProjectId != "P1" || state.VersionId != "V1" {
		t.Errorf("Slug, ProjectId, VersionId = %q, %q, %q", state.Slug, state.ProjectId, state.VersionId)
	}
}

func TestFromArchiveFailedLookupLeavesIdsEmpty(t *testing.T) {
	t.Parallel()
	pack := writePack(t, mrpack.Index{Name: "Example"}, nil)

	state, err := FromArchive(pack, Options{}, noIds)
	if err != nil {
		t.Fatal(err)
	}

	if state.Slug != "" || state.ProjectId != "" || state.VersionId != "" {
		t.Errorf("Slug, ProjectId, VersionId = %q, %q, %q, want empty", state.Slug, state.ProjectId, state.VersionId)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	state := &Schema{
		SchemaVersion: 1,
		Name:          "Example",
		Version:       "1.2.0",
		ArchiveSha512: "abc",
		Ids:           Ids{Slug: "example"},
		Options:       Options{OptionalSelect: []string{"opt-a"}},
		Deps:          mrpack.Deps{Minecraft: "1.21.8"},
		Hashes:        map[string]modrinth.Hashes{"mods/a.jar": {Sha1: "1", Sha512: "512"}},
	}

	if err := state.Save(dir); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPackState(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, state) {
		t.Errorf("LoadPackState() = %+v, want %+v", got, state)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != FileName {
		t.Errorf("server dir holds %v, want only %s", entries, FileName)
	}
}

func TestLoadWithoutState(t *testing.T) {
	t.Parallel()
	state, err := LoadPackState(t.TempDir())
	if err != nil || state != nil {
		t.Errorf("LoadPackState() = %+v, %v, want nil, nil", state, err)
	}
}

func writeState(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadUnknownSchemaFails(t *testing.T) {
	t.Parallel()
	dir := writeState(t, `{"schema":2}`)

	state, err := LoadPackState(dir)
	if err == nil {
		t.Errorf("LoadPackState() = %+v, nil, want an error", state)
	}
}

func TestExcludes(t *testing.T) {
	t.Parallel()
	opts := Options{Exclude: []string{"config/**", "mods/*-client.jar"}}
	tests := map[string]bool{
		"config/a.txt":          true,
		"config/sub/dir/b.json": true,
		"mods/foo-client.jar":   true,
		"mods/foo.jar":          false,
		"mods/sub/x-client.jar": false,
		"configs/a.txt":         false,
		"other/config/a.txt":    false,
	}
	for p, want := range tests {
		if got := opts.Excludes(p); got != want {
			t.Errorf("Excludes(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestExcludesDoubleStarInTheMiddle(t *testing.T) {
	t.Parallel()
	opts := Options{Exclude: []string{"world/**/region/*.mca"}}
	tests := map[string]bool{
		"world/region/r.0.0.mca":        true,
		"world/DIM-1/region/r.0.0.mca":  true,
		"world/a/b/region/r.0.0.mca":    true,
		"world/region/r.0.0.mca.bak":    false,
		"world/DIM-1/regions/r.0.0.mca": false,
	}
	for p, want := range tests {
		if got := opts.Excludes(p); got != want {
			t.Errorf("Excludes(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestExcludesLeadingDoubleStar(t *testing.T) {
	t.Parallel()
	opts := Options{Exclude: []string{"**/*-client.jar"}}
	tests := map[string]bool{
		"a-client.jar":        true,
		"mods/a-client.jar":   true,
		"mods/x/a-client.jar": true,
		"mods/a.jar":          false,
	}
	for p, want := range tests {
		if got := opts.Excludes(p); got != want {
			t.Errorf("Excludes(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"config/**", "mods/*-client.jar", "a/[bc]/d"} {
		if err := CheckPattern(p); err != nil {
			t.Errorf("CheckPattern(%q) = %v, want nil", p, err)
		}
	}
	for _, p := range []string{"mods/[a.jar", "", "/abs/path", "config/", "./config/**", "config//x", "a/../b"} {
		if err := CheckPattern(p); err == nil {
			t.Errorf("CheckPattern(%q) = nil, want an error", p)
		}
	}
}

func TestShippedFilesSkipsExcluded(t *testing.T) {
	t.Parallel()
	index := mrpack.Index{Name: "Example", Files: []mrpack.File{
		indexFile("mods/required.jar", modrinth.RequiredEnvSupport, "required"),
		indexFile("mods/foo-client.jar", modrinth.RequiredEnvSupport, "client"),
	}}
	pack := writePack(t, index, [][2]string{{"overrides/config/a.txt", "server"}})

	got := ShippedFiles(&index, pack, Options{Exclude: []string{"config/**", "mods/*-client.jar"}})

	want := map[string]modrinth.Hashes{"mods/required.jar": {Sha1: "required-sha1", Sha512: "required-sha512"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ShippedFiles() = %v, want %v", got, want)
	}
}

func TestInvalidateDropsArchiveHash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	state := &Schema{SchemaVersion: 1, Name: "Example", ArchiveSha512: "abc", Hashes: map[string]modrinth.Hashes{}}

	if err := state.Invalidate(dir); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPackState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.ArchiveSha512 != "" || got.Name != "Example" {
		t.Errorf("saved state = %+v, want Example without archive hash", got)
	}
	if state.ArchiveSha512 != "abc" {
		t.Errorf("in-memory state lost its archive hash")
	}
}

func TestArchiveSha512(t *testing.T) {
	t.Parallel()
	pack := writePack(t, mrpack.Index{Name: "Example"}, nil)

	got, err := ArchiveSha512(pack)
	if err != nil {
		t.Fatal(err)
	}

	if want := archiveSha512(t, pack); got != want {
		t.Errorf("ArchiveSha512() = %q, want %q", got, want)
	}
}

func TestForceSyncIsNotRecorded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	state := &Schema{SchemaVersion: 1, Options: Options{ForceSync: true}, Hashes: map[string]modrinth.Hashes{}}
	if err := state.Save(dir); err != nil {
		t.Fatal(err)
	}

	got, err := LoadPackState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Options.ForceSync {
		t.Error("ForceSync was recorded")
	}
	if !got.Options.Equal(Options{ForceSync: true}) {
		t.Error("Equal() compares ForceSync, want it ignored")
	}
}

// copyBeta lays out a server dir holding the beta fixture as its state.
func copyBeta(t *testing.T, files map[string]string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/beta-packstate.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), b, 0644); err != nil {
		t.Fatal(err)
	}
	for p, data := range files {
		if err := os.WriteFile(filepath.Join(dir, p), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestLoadBetaState(t *testing.T) {
	t.Parallel()
	dir := copyBeta(t, map[string]string{"fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.1.0.jar": "launcher"})

	state, err := LoadPackState(dir)
	if err != nil {
		t.Fatal(err)
	}

	if state == nil {
		t.Fatal("LoadPackState() = nil, want the converted beta state")
	}
	if state.SchemaVersion != 1 || state.Version != "1.0.0" || state.ArchiveSha512 != "" {
		t.Errorf("SchemaVersion, Version, ArchiveSha512 = %d, %q, %q", state.SchemaVersion, state.Version, state.ArchiveSha512)
	}
	if state.Ids != (Ids{Slug: "example", ProjectId: "P1", VersionId: "V1"}) {
		t.Errorf("Ids = %+v", state.Ids)
	}
	if state.Deps != (mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.17.2"}) {
		t.Errorf("Deps = %+v", state.Deps)
	}
	// The beta override bug recorded the SHA-512 of empty input, so the
	// override compares by SHA-1 only.
	want := map[string]modrinth.Hashes{
		"mods/a.jar":   {Sha1: "f92e777f4341930bad9b2422283c4680d00dbc06", Sha512: "483365784455682e7e83c6e150e56dbe3f387a06c9d407b8253e4bdf8e0bd1fac8234e3973d4e59ff8c01cb969f25c0d08c5dc5bb967744b1a5813afe3c58fcc"},
		"config/a.txt": {Sha1: "356a192b7913b04c54574d18c28d46e6395428ab"},
	}
	if !reflect.DeepEqual(state.Hashes, want) {
		t.Errorf("Hashes = %v, want %v", state.Hashes, want)
	}
	if state.ServerFile == nil || state.ServerFile.Path != "fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.1.0.jar" || state.ServerFile.Hashes.Sha512 == "" {
		t.Errorf("ServerFile = %+v, want the fabric launcher with its hashes", state.ServerFile)
	}
}

func TestInferServerFile(t *testing.T) {
	t.Parallel()
	fabric := mrpack.Deps{Minecraft: "1.21.8", Fabric: "0.17.2"}
	tests := []struct {
		name  string
		deps  mrpack.Deps
		files []string
		want  string
	}{
		{"fabric", fabric, []string{"fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.1.0.jar", "server.jar"}, "fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.1.0.jar"},
		{"two fabric launchers", fabric, []string{"fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.0.3.jar", "fabric-server-mc.1.21.8-loader.0.17.2-launcher.1.1.0.jar"}, ""},
		{"other fabric loader", fabric, []string{"fabric-server-mc.1.21.8-loader.0.16.14-launcher.1.1.0.jar"}, ""},
		{"forge", mrpack.Deps{Minecraft: "1.16.5", Forge: "36.2.39"}, []string{"forge-1.16.5-36.2.39.jar", "forge-1.16.5-36.2.39-installer.jar"}, "forge-1.16.5-36.2.39.jar"},
		{"forge universal", mrpack.Deps{Minecraft: "1.12.2", Forge: "14.23.5.2768"}, []string{"forge-1.12.2-14.23.5.2768-universal.jar"}, "forge-1.12.2-14.23.5.2768-universal.jar"},
		{"forge shim", mrpack.Deps{Minecraft: "1.20.4", Forge: "49.1.0"}, []string{"forge-1.20.4-49.1.0-shim.jar", "run.sh"}, "forge-1.20.4-49.1.0-shim.jar"},
		{"forge run.sh only", mrpack.Deps{Minecraft: "1.20.1", Forge: "47.3.0"}, []string{"run.sh"}, ""},
		{"vanilla", mrpack.Deps{Minecraft: "1.21.8"}, []string{"server.jar"}, ""},
		{"nothing installed", fabric, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := inferServerFile(dir, tt.deps)
			if err != nil {
				t.Fatal(err)
			}

			if tt.want == "" {
				if got != nil {
					t.Errorf("inferServerFile() = %+v, want nil", got)
				}
				return
			}
			if got == nil || got.Path != tt.want || got.Hashes.Sha1 == "" || got.Hashes.Sha512 == "" {
				t.Errorf("inferServerFile() = %+v, want %s with its hashes", got, tt.want)
			}
		})
	}
}
