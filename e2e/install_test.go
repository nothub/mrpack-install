//go:build e2e

package e2e

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nothub/mrpack-install/modrinth/api"
)

// packIndex is modrinth.index.json as the Modrinth pack format documents it.
type packIndex struct {
	FormatVersion int               `json:"formatVersion"`
	Game          string            `json:"game"`
	VersionID     string            `json:"versionId"`
	Name          string            `json:"name"`
	Files         []packFile        `json:"files"`
	Dependencies  map[string]string `json:"dependencies"`
}

type packFile struct {
	Path      string            `json:"path"`
	Hashes    map[string]string `json:"hashes"`
	Env       map[string]string `json:"env"`
	Downloads []string          `json:"downloads"`
	FileSize  int               `json:"fileSize"`
}

// zipEntry is one file in a pack archive, written in slice order.
type zipEntry struct {
	Name string
	Data string
}

// packArchive returns an mrpack holding index followed by entries.
func packArchive(t testing.TB, index packIndex, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("modrinth.index.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(w).Encode(index); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		w, err := zw.Create(e.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.Data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// seedPackMod seeds a mod named slug and returns its jar as a pack file at
// mods/<slug>.jar, downloaded from the cdn.
func seedPackMod(t testing.TB, slug string) (packFile, []byte) {
	t.Helper()
	project := seedMod(t, slug, fabric1218)
	version, err := api.NewClient(labrinthURL).GetVersion(project.Versions[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(version.Files) != 1 {
		t.Fatalf("files = %v, want one", version.Files)
	}
	jar := fabricJar(t, slug)
	sum1 := sha1.Sum(jar)
	sum512 := sha512.Sum512(jar)
	return packFile{
		Path: "mods/" + slug + ".jar",
		Hashes: map[string]string{
			"sha1":   hex.EncodeToString(sum1[:]),
			"sha512": hex.EncodeToString(sum512[:]),
		},
		Env:       map[string]string{"client": "required", "server": "required"},
		Downloads: []string{version.Files[0].Url},
		FileSize:  len(jar),
	}, jar
}

// install runs mrpack-install for slug against labrinth and returns the
// server dir it installed into.
func install(t testing.TB, slug string) string {
	t.Helper()
	dir := t.TempDir()
	installInto(t, dir, slug)
	return dir
}

// installInto runs mrpack-install with args against labrinth into dir.
func installInto(t testing.TB, dir string, args ...string) {
	t.Helper()
	args = append(args, "--host", labrinthURL, "--server-dir", dir)
	out, err := mrpackInstall(t, args...)
	if err != nil {
		t.Fatalf("mrpack-install %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	t.Logf("mrpack-install %s:\n%s", strings.Join(args, " "), out)
}

// backupRuns returns the per-run dirs under the backup base.
func backupRuns(t testing.TB, base string) []string {
	t.Helper()
	runs, err := filepath.Glob(filepath.Join(base, "*"))
	if err != nil {
		t.Fatal(err)
	}
	return runs
}

func assertMissing(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s: stat error = %v, want not exist", path, err)
	}
}

// packState is the part of packstate.json the e2e tests check.
type packState struct {
	Schema        int    `json:"schema"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	ArchiveSha512 string `json:"archive-sha512"`
	Slug          string `json:"slug"`
	ProjectID     string `json:"project-id"`
	VersionID     string `json:"version-id"`
	Options       struct {
		Exclude []string `json:"exclude"`
	} `json:"options"`
	Hashes map[string]map[string]string `json:"hashes"`
}

func readState(t testing.TB, dir string) packState {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "packstate.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state packState
	if err := json.Unmarshal(b, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func assertContent(t testing.TB, path string, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func assertSha512(t testing.TB, path string, data []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(got, data) {
		t.Errorf("%s: %d bytes differ from the %d uploaded", path, len(got), len(data))
	}
}

func TestInstallFabricPack(t *testing.T) {
	t.Parallel()
	const slug = "e2e-install-fabric"
	mod, jar := seedPackMod(t, slug+"-mod")
	pack := packArchive(t, packIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     "1.0.0",
		Name:          slug,
		Files:         []packFile{mod},
		Dependencies:  map[string]string{"minecraft": "1.21.8", "fabric-loader": "0.17.2"},
	}, []zipEntry{
		// server-overrides/ comes first in the archive, so the shared
		// path also checks that precedence ignores entry order.
		{"server-overrides/config/shared.txt", "server"},
		{"overrides/config/shared.txt", "common"},
		{"overrides/config/common-only.txt", "common"},
		{"client-overrides/config/client-only.txt", "client"},
	})
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)

	dir := install(t, slug)

	jars, err := filepath.Glob(filepath.Join(dir, "fabric-server-mc.1.21.8-loader.0.17.2-launcher.*.jar"))
	if err != nil {
		t.Fatal(err)
	}
	if len(jars) != 1 {
		t.Errorf("fabric server jars = %v, want one", jars)
	}
	assertSha512(t, filepath.Join(dir, mod.Path), jar)
	assertContent(t, filepath.Join(dir, "config", "common-only.txt"), "common")
	assertContent(t, filepath.Join(dir, "config", "shared.txt"), "server")
	if _, err := os.Stat(filepath.Join(dir, "config", "client-only.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("client override: stat error = %v, want not exist", err)
	}
	// A file input gets its ids from the archive hash lookup.
	path := filepath.Join(t.TempDir(), slug+".mrpack")
	if err := os.WriteFile(path, pack, 0644); err != nil {
		t.Fatal(err)
	}
	fileDir := install(t, path)
	for _, d := range []string{dir, fileDir} {
		state := readState(t, d)
		if state.Slug != slug || state.ProjectID != created.Id || state.VersionID != created.Versions[0] {
			t.Errorf("%s: slug, project-id, version-id = %q, %q, %q, want %q, %q, %q", d,
				state.Slug, state.ProjectID, state.VersionID, slug, created.Id, created.Versions[0])
		}
	}
}

func TestInstallVanillaPack(t *testing.T) {
	t.Parallel()
	const slug = "e2e-install-vanilla"
	mod, jar := seedPackMod(t, slug+"-mod")
	pack := packArchive(t, packIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     "1.0.0",
		Name:          slug,
		Files:         []packFile{mod},
		Dependencies:  map[string]string{"minecraft": "1.21.8"},
	}, []zipEntry{
		{"overrides/config/common.txt", "common"},
		{"server-overrides/config/server.txt", "server"},
	})
	// labrinth takes only mod loaders as a pack's loaders. mrpack-install
	// reads the index dependencies, so the fabric tag changes nothing.
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)

	dir := install(t, slug)

	if fi, err := os.Stat(filepath.Join(dir, "server.jar")); err != nil {
		t.Error(err)
	} else if fi.Size() == 0 {
		t.Error("server.jar is empty")
	}
	assertSha512(t, filepath.Join(dir, mod.Path), jar)
	assertContent(t, filepath.Join(dir, "config", "common.txt"), "common")
	assertContent(t, filepath.Join(dir, "config", "server.txt"), "server")
}

func TestInstallFailsOnMissingDownload(t *testing.T) {
	t.Parallel()
	const slug = "e2e-install-missing-dl"
	mod, _ := seedPackMod(t, slug+"-mod")
	missing := mod
	missing.Path = "mods/" + slug + "-missing.jar"
	missing.Downloads = []string{mod.Downloads[0] + ".missing"}
	pack := packArchive(t, packIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     "1.0.0",
		Name:          slug,
		Files:         []packFile{mod, missing},
		Dependencies:  map[string]string{"minecraft": "1.21.8", "fabric-loader": "0.17.2"},
	}, nil)
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)

	dir := t.TempDir()
	out, err := mrpackInstall(t, slug, "--host", labrinthURL, "--server-dir", dir)

	if err == nil {
		t.Fatalf("mrpack-install %s: exit 0, want non-zero\n%s", slug, out)
	}
	// Per-attempt log lines also name the file, so check only the last
	// line, which carries the returned error.
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if last := lines[len(lines)-1]; !strings.Contains(last, missing.Path) || strings.Contains(last, mod.Path) {
		t.Errorf("last line = %q, want it to name only %s", last, missing.Path)
	}
	if _, err := os.Stat(filepath.Join(dir, "packstate.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("packstate.json: stat error = %v, want not exist", err)
	}
}

// A pack labrinth doesn't know still gets a state, without ids.
func TestInstallLocalPackWritesState(t *testing.T) {
	t.Parallel()
	const slug = "e2e-install-local"
	mod, _ := seedPackMod(t, slug+"-mod")
	pack := packArchive(t, packIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     "1.0.0",
		Name:          slug,
		Files:         []packFile{mod},
		Dependencies:  map[string]string{"minecraft": "1.21.8", "fabric-loader": "0.17.2"},
	}, []zipEntry{{"overrides/config/common.txt", "common"}})
	path := filepath.Join(t.TempDir(), slug+".mrpack")
	if err := os.WriteFile(path, pack, 0644); err != nil {
		t.Fatal(err)
	}

	dir := install(t, path)

	state := readState(t, dir)
	if state.Schema != 1 {
		t.Errorf("schema = %d, want 1", state.Schema)
	}
	if state.Name != slug || state.Version != "1.0.0" {
		t.Errorf("name, version = %q, %q, want %q, %q", state.Name, state.Version, slug, "1.0.0")
	}
	sum := sha512.Sum512(pack)
	if want := hex.EncodeToString(sum[:]); state.ArchiveSha512 != want {
		t.Errorf("archive-sha512 = %q, want %q", state.ArchiveSha512, want)
	}
	if state.Slug != "" || state.ProjectID != "" || state.VersionID != "" {
		t.Errorf("slug, project-id, version-id = %q, %q, %q, want empty", state.Slug, state.ProjectID, state.VersionID)
	}
	if got := state.Hashes[mod.Path]["sha512"]; got != mod.Hashes["sha512"] {
		t.Errorf("hashes[%s].sha512 = %q, want %q", mod.Path, got, mod.Hashes["sha512"])
	}
	if _, ok := state.Hashes["config/common.txt"]; !ok {
		t.Errorf("hashes = %v, want config/common.txt", state.Hashes)
	}
}

var fabricDeps = map[string]string{"minecraft": "1.21.8", "fabric-loader": "0.17.2"}

func fabricPack(t testing.TB, name string, version string, files []packFile, entries []zipEntry) []byte {
	t.Helper()
	return packArchive(t, packIndex{
		FormatVersion: 1,
		Game:          "minecraft",
		VersionID:     version,
		Name:          name,
		Files:         files,
		Dependencies:  fabricDeps,
	}, entries)
}

// Each install of a newer version deletes what the version dropped and
// leaves files no pack version listed alone.
func TestInstallSyncsToNewVersion(t *testing.T) {
	t.Parallel()
	const slug = "e2e-sync-versions"
	keep, keepJar := seedPackMod(t, slug+"-keep")
	drop, _ := seedPackMod(t, slug+"-drop")
	v1 := fabricPack(t, slug, "1.0.0", []packFile{keep, drop}, []zipEntry{
		{"overrides/config/all.txt", "1"},
		{"overrides/config/v1/only.txt", "1"},
		{"overrides/config/tuned.txt", "1"},
		{"overrides/config/tuned-dropped.txt", "1"},
	})
	v2 := fabricPack(t, slug, "2.0.0", []packFile{keep}, []zipEntry{
		{"overrides/config/all.txt", "2"},
		{"overrides/config/tuned.txt", "2"},
	})
	v3 := fabricPack(t, slug, "3.0.0", []packFile{}, []zipEntry{{"overrides/config/all.txt", "3"}})
	version := func(number string) versionData {
		v := fabric1218
		v.VersionNumber = number
		return v
	}
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + "-1.0.0.mrpack", Data: v1}, version("1.0.0"))
	uploadVersion(t, created.Id, file{Name: slug + "-2.0.0.mrpack", Data: v2}, version("2.0.0"))
	uploadVersion(t, created.Id, file{Name: slug + "-3.0.0.mrpack", Data: v3}, version("3.0.0"))
	approve(t, created.Id)
	dir := t.TempDir()
	// What a killed run would leave behind: a dir where this run stages
	// a download, which fails the run unless the leftover is removed first.
	leftover := filepath.Join(dir, ".mrpack-install-staging", filepath.FromSlash(keep.Path), "partial")
	if err := os.MkdirAll(leftover, 0755); err != nil {
		t.Fatal(err)
	}

	installInto(t, dir, slug, "1.0.0")
	assertMissing(t, filepath.Join(dir, ".mrpack-install-staging"))
	assertSha512(t, filepath.Join(dir, keep.Path), keepJar)
	assertContent(t, filepath.Join(dir, "config", "v1", "only.txt"), "1")
	mine := filepath.Join(dir, "mods", "mine.jar")
	if err := os.WriteFile(mine, []byte("mine"), 0644); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "world", "empty")
	if err := os.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}
	// Edited files are backed up before v2 overwrites or deletes them.
	tuned := filepath.Join(dir, "config", "tuned.txt")
	tunedDropped := filepath.Join(dir, "config", "tuned-dropped.txt")
	for _, p := range []string{tuned, tunedDropped} {
		if err := os.WriteFile(p, []byte("mine"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	installInto(t, dir, slug, "2.0.0")
	assertSha512(t, filepath.Join(dir, keep.Path), keepJar)
	assertMissing(t, filepath.Join(dir, drop.Path))
	assertMissing(t, filepath.Join(dir, "config", "v1"))
	assertContent(t, filepath.Join(dir, "config", "all.txt"), "2")
	assertContent(t, mine, "mine")
	assertContent(t, tuned, "2")
	assertMissing(t, tunedDropped)
	runs := backupRuns(t, filepath.Join(dir, "backups"))
	if len(runs) != 1 {
		t.Fatalf("backup runs = %v, want one", runs)
	}
	assertContent(t, filepath.Join(runs[0], "config", "tuned.txt"), "mine")
	assertContent(t, filepath.Join(runs[0], "config", "tuned-dropped.txt"), "mine")
	for _, unedited := range []string{"config/all.txt", "config/v1/only.txt", drop.Path} {
		assertMissing(t, filepath.Join(runs[0], unedited))
	}

	installInto(t, dir, slug, "3.0.0")
	assertMissing(t, filepath.Join(dir, keep.Path))
	assertContent(t, filepath.Join(dir, "config", "all.txt"), "3")
	assertContent(t, mine, "mine")
	if _, err := os.Stat(empty); err != nil {
		t.Errorf("admin's empty dir: %v", err)
	}
	// v3 replaces and drops only unedited files, so it backs up nothing.
	if runs := backupRuns(t, filepath.Join(dir, "backups")); len(runs) != 1 {
		t.Errorf("backup runs = %v, want one", runs)
	}
	assertMissing(t, filepath.Join(dir, ".mrpack-install-staging"))
	if got := readState(t, dir).Version; got != "3.0.0" {
		t.Errorf("state version = %q, want 3.0.0", got)
	}
}

func TestInstallSyncsLocalPacks(t *testing.T) {
	t.Parallel()
	const slug = "e2e-sync-local"
	mod, jar := seedPackMod(t, slug+"-mod")
	packs := t.TempDir()
	v1 := filepath.Join(packs, "v1.mrpack")
	v2 := filepath.Join(packs, "v2.mrpack")
	if err := os.WriteFile(v1, fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{{"overrides/config/dropped.txt", "1"}}), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(v2, fabricPack(t, slug, "2.0.0", []packFile{mod}, []zipEntry{{"overrides/config/kept.txt", "2"}}), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	backups := t.TempDir()

	installInto(t, dir, v1)
	dropped := filepath.Join(dir, "config", "dropped.txt")
	assertContent(t, dropped, "1")
	if err := os.WriteFile(dropped, []byte("mine"), 0644); err != nil {
		t.Fatal(err)
	}
	launchers, err := filepath.Glob(filepath.Join(dir, "fabric-server-*.jar"))
	if err != nil || len(launchers) != 1 {
		t.Fatalf("fabric server jars = %v, %v, want one", launchers, err)
	}
	before, err := os.Stat(launchers[0])
	if err != nil {
		t.Fatal(err)
	}
	installInto(t, dir, v2, "--backup-dir", backups)

	// v2 has v1's dependencies, so the server is not installed again.
	after, err := os.Stat(launchers[0])
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("%s was replaced or modified", launchers[0])
	}
	assertMissing(t, dropped)

	// A deleted server jar comes back on the next run.
	if err := os.Remove(launchers[0]); err != nil {
		t.Fatal(err)
	}
	installInto(t, dir, v2, "--backup-dir", backups)
	if _, err := os.Stat(launchers[0]); err != nil {
		t.Errorf("deleted server jar not reinstalled: %v", err)
	}
	// A new --server-file installs the server under that name and drops
	// the old jar.
	installInto(t, dir, v2, "--backup-dir", backups, "--server-file", "mc.jar")
	if _, err := os.Stat(filepath.Join(dir, "mc.jar")); err != nil {
		t.Errorf("--server-file mc.jar: %v", err)
	}
	assertMissing(t, launchers[0])

	runs := backupRuns(t, backups)
	if len(runs) != 1 {
		t.Fatalf("backup runs = %v, want one", runs)
	}
	assertContent(t, filepath.Join(runs[0], "config", "dropped.txt"), "mine")
	assertContent(t, filepath.Join(dir, "config", "kept.txt"), "2")
	assertSha512(t, filepath.Join(dir, mod.Path), jar)
}

// The archive of a slug install is cached outside the server dir, and a
// second install of the same version reuses it.
func TestInstallCachesArchive(t *testing.T) {
	t.Parallel()
	const slug = "e2e-install-cache"
	mod, _ := seedPackMod(t, slug+"-mod")
	pack := fabricPack(t, slug, "1.0.0", []packFile{mod}, nil)
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)
	sum := sha512.Sum512(pack)
	cached := filepath.Join(cacheHome, "mrpack-install", hex.EncodeToString(sum[:])+".mrpack")
	// Two server dirs, so the second run can only skip the download
	// through the cache.
	dirs := []string{t.TempDir(), t.TempDir()}

	installInto(t, dirs[0], slug)
	before, err := os.Stat(cached)
	if err != nil {
		t.Fatal(err)
	}
	installInto(t, dirs[1], slug)

	after, err := os.Stat(cached)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("cached archive was replaced or modified by the second run")
	}
	for _, dir := range dirs {
		archives, err := filepath.Glob(filepath.Join(dir, "*.mrpack"))
		if err != nil {
			t.Fatal(err)
		}
		if len(archives) != 0 {
			t.Errorf("server dir holds archives %v, want none", archives)
		}
	}
}

// snapshot maps every file under dir to its content, relative to dir.
func snapshot(t testing.TB, dir string) map[string]string {
	t.Helper()
	files := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		files[rel] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// A failed download leaves the server dir, the backups and the state as
// they were before the run.
func TestSyncFailedDownloadChangesNothing(t *testing.T) {
	t.Parallel()
	const slug = "e2e-sync-failed-dl"
	mod, _ := seedPackMod(t, slug+"-mod")
	missing := mod
	missing.Path = "mods/" + slug + "-missing.jar"
	missing.Downloads = []string{mod.Downloads[0] + ".missing"}
	packs := t.TempDir()
	v1 := filepath.Join(packs, "v1.mrpack")
	v2 := filepath.Join(packs, "v2.mrpack")
	if err := os.WriteFile(v1, fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{
		{"overrides/config/tuned.txt", "1"},
		{"overrides/config/dropped.txt", "1"},
	}), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(v2, fabricPack(t, slug, "2.0.0", []packFile{mod, missing}, []zipEntry{
		{"overrides/config/tuned.txt", "2"},
	}), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	installInto(t, dir, v1)
	if err := os.WriteFile(filepath.Join(dir, "config", "tuned.txt"), []byte("mine"), 0644); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, dir)

	out, err := mrpackInstall(t, v2, "--host", labrinthURL, "--server-dir", dir)

	if err == nil {
		t.Fatalf("mrpack-install v2: exit 0, want non-zero\n%s", out)
	}
	after := snapshot(t, dir)
	if !reflect.DeepEqual(after, before) {
		for p := range after {
			if after[p] != before[p] {
				t.Errorf("%s changed or appeared", p)
			}
		}
		for p := range before {
			if _, ok := after[p]; !ok {
				t.Errorf("%s disappeared", p)
			}
		}
	}
	assertMissing(t, filepath.Join(dir, "backups"))
	assertMissing(t, filepath.Join(dir, ".mrpack-install-staging"))
}

// Excluded paths keep v1's content, also when v2 changes or drops them.
func TestSyncSkipsExcludedPaths(t *testing.T) {
	t.Parallel()
	const slug = "e2e-sync-exclude"
	mod, _ := seedPackMod(t, slug+"-mod")
	packs := t.TempDir()
	v1 := filepath.Join(packs, "v1.mrpack")
	v2 := filepath.Join(packs, "v2.mrpack")
	if err := os.WriteFile(v1, fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{
		{"overrides/config/tuned/a.txt", "1"},
		{"overrides/config/dropped.txt", "1"},
		{"overrides/config/normal.txt", "1"},
	}), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(v2, fabricPack(t, slug, "2.0.0", []packFile{mod}, []zipEntry{
		{"overrides/config/tuned/a.txt", "2"},
		{"overrides/config/normal.txt", "2"},
	}), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	installInto(t, dir, v1)

	installInto(t, dir, v2, "--exclude", "config/tuned/**", "--exclude", "config/dropped.*")

	assertContent(t, filepath.Join(dir, "config", "tuned", "a.txt"), "1")
	assertContent(t, filepath.Join(dir, "config", "dropped.txt"), "1")
	assertContent(t, filepath.Join(dir, "config", "normal.txt"), "2")
	assertMissing(t, filepath.Join(dir, "backups"))
	state := readState(t, dir)
	for _, p := range []string{"config/tuned/a.txt", "config/dropped.txt"} {
		if _, ok := state.Hashes[p]; ok {
			t.Errorf("state records excluded %s", p)
		}
	}
	if want := []string{"config/tuned/**", "config/dropped.*"}; !reflect.DeepEqual(state.Options.Exclude, want) {
		t.Errorf("state options exclude = %v, want %v", state.Options.Exclude, want)
	}
}

// A loader bump installs the new server and deletes the old loader's jar.
func TestSyncReinstallsServerOnLoaderBump(t *testing.T) {
	t.Parallel()
	const slug = "e2e-sync-loader-bump"
	mod, _ := seedPackMod(t, slug+"-mod")
	pack := func(version string, loader string) string {
		data := packArchive(t, packIndex{
			FormatVersion: 1,
			Game:          "minecraft",
			VersionID:     version,
			Name:          slug,
			Files:         []packFile{mod},
			Dependencies:  map[string]string{"minecraft": "1.21.8", "fabric-loader": loader},
		}, nil)
		path := filepath.Join(t.TempDir(), version+".mrpack")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	dir := t.TempDir()

	installInto(t, dir, pack("1.0.0", "0.16.14"))
	installInto(t, dir, pack("2.0.0", "0.17.2"))

	jars, err := filepath.Glob(filepath.Join(dir, "fabric-server-*.jar"))
	if err != nil {
		t.Fatal(err)
	}
	if len(jars) != 1 || !strings.Contains(filepath.Base(jars[0]), "loader.0.17.2-") {
		t.Errorf("fabric server jars = %v, want only the 0.17.2 one", jars)
	}
	assertMissing(t, filepath.Join(dir, "backups"))
}

// modTimes maps every file under dir to its modification time.
func modTimes(t testing.TB, dir string) map[string]time.Time {
	t.Helper()
	times := make(map[string]time.Time)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		times[path] = fi.ModTime()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return times
}

// A re-run for the installed archive changes no file and, for a slug,
// downloads nothing.
func TestRerunIsNoOp(t *testing.T) {
	t.Parallel()
	const slug = "e2e-rerun-noop"
	mod, _ := seedPackMod(t, slug+"-mod")
	pack := fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{{"overrides/config/a.txt", "1"}})
	created := createProject(t, api.ModpackProjectType, slug, file{Name: slug + ".mrpack", Data: pack}, fabric1218)
	approve(t, created.Id)
	sum := sha512.Sum512(pack)
	cached := filepath.Join(cacheHome, "mrpack-install", hex.EncodeToString(sum[:])+".mrpack")
	dir := t.TempDir()
	installInto(t, dir, slug)
	contents, times := snapshot(t, dir), modTimes(t, dir)
	if err := os.Remove(cached); err != nil {
		t.Fatal(err)
	}

	installInto(t, dir, slug)

	if !reflect.DeepEqual(snapshot(t, dir), contents) {
		t.Error("a file changed")
	}
	if !reflect.DeepEqual(modTimes(t, dir), times) {
		t.Error("a modification time changed")
	}
	assertMissing(t, cached)
}

// --force-sync restores a deleted pack file, and backs up and resets an
// edited one.
func TestForceSyncRepairs(t *testing.T) {
	t.Parallel()
	const slug = "e2e-force-sync"
	mod, jar := seedPackMod(t, slug+"-mod")
	path := filepath.Join(t.TempDir(), "pack.mrpack")
	if err := os.WriteFile(path, fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{{"overrides/config/a.txt", "1"}}), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	installInto(t, dir, path)
	// A re-run of a local archive is a no-op too.
	times := modTimes(t, dir)
	installInto(t, dir, path)
	if !reflect.DeepEqual(modTimes(t, dir), times) {
		t.Error("re-run of a local archive changed a modification time")
	}
	launchers, err := filepath.Glob(filepath.Join(dir, "fabric-server-*.jar"))
	if err != nil || len(launchers) != 1 {
		t.Fatalf("fabric server jars = %v, %v, want one", launchers, err)
	}
	before, err := os.Stat(launchers[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, mod.Path)); err != nil {
		t.Fatal(err)
	}
	edited := filepath.Join(dir, "config", "a.txt")
	if err := os.WriteFile(edited, []byte("mine"), 0644); err != nil {
		t.Fatal(err)
	}

	installInto(t, dir, path, "--force-sync")

	after, err := os.Stat(launchers[0])
	if err != nil {
		t.Fatal(err)
	}
	if after.ModTime().Equal(before.ModTime()) {
		t.Errorf("--force-sync did not reinstall the server: %s unchanged", launchers[0])
	}
	assertSha512(t, filepath.Join(dir, mod.Path), jar)
	assertContent(t, edited, "1")
	runs := backupRuns(t, filepath.Join(dir, "backups"))
	if len(runs) != 1 {
		t.Fatalf("backup runs = %v, want one", runs)
	}
	assertContent(t, filepath.Join(runs[0], "config", "a.txt"), "mine")
}

// --dry-run prints the plan and changes nothing, not even a missing dir.
func TestDryRun(t *testing.T) {
	t.Parallel()
	const slug = "e2e-dry-run"
	mod, _ := seedPackMod(t, slug+"-mod")
	packs := t.TempDir()
	v1 := filepath.Join(packs, "v1.mrpack")
	v2 := filepath.Join(packs, "v2.mrpack")
	if err := os.WriteFile(v1, fabricPack(t, slug, "1.0.0", []packFile{mod}, []zipEntry{{"overrides/config/dropped.txt", "1"}}), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(v2, fabricPack(t, slug, "2.0.0", []packFile{mod}, []zipEntry{{"overrides/config/new.txt", "2"}}), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	installInto(t, dir, v1)
	contents, times := snapshot(t, dir), modTimes(t, dir)

	stdout, stderr, err := mrpackInstallStdout(t, v2, "--dry-run", "--host", labrinthURL, "--server-dir", dir)
	if err != nil {
		t.Fatalf("mrpack-install --dry-run: %v\n%s", err, stderr)
	}

	if want := "delete config/dropped.txt\nwrite config/new.txt\ninstall server: no\n"; string(stdout) != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if !reflect.DeepEqual(snapshot(t, dir), contents) || !reflect.DeepEqual(modTimes(t, dir), times) {
		t.Error("--dry-run changed the server dir")
	}

	missing := filepath.Join(t.TempDir(), "new-server")
	stdout, stderr, err = mrpackInstallStdout(t, v1, "--dry-run", "--host", labrinthURL, "--server-dir", missing)
	if err != nil {
		t.Fatalf("mrpack-install --dry-run: %v\n%s", err, stderr)
	}
	if !strings.Contains(string(stdout), "install server: yes\n") {
		t.Errorf("stdout lacks the server install:\n%s", stdout)
	}
	assertMissing(t, missing)
}
