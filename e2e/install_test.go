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
	"os/exec"
	"path/filepath"
	"testing"

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
	cmd := exec.CommandContext(t.Context(), binary, slug, "--host", labrinthURL, "--server-dir", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mrpack-install %s: %v\n%s", slug, err, out)
	}
	// A failed file download is only logged, and the exit code stays 0.
	t.Logf("mrpack-install %s:\n%s", slug, out)
	return dir
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
