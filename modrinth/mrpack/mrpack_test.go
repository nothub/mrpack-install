package mrpack

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
)

const examplePack = "testdata/example.mrpack"

// hashes of the single newline byte every override file in the example pack holds
const newlineSha1 = "adc83b19e793491b1c6ea0fd8b46cd9f32e592fc"
const newlineSha512 = "be688838ca8686e5c90689bf2ab585cef1137c999b48c70b92f67a5c34dc15697b5d11c982ed6d71be1e1e7f7b4e0733884aa97c3f7a339a8ed03577cf74be09"

func TestReadIndex(t *testing.T) {
	t.Parallel()
	index, err := ReadIndex(examplePack)
	if err != nil {
		t.Fatal(err)
	}
	if index.Format != 1 {
		t.Errorf("Format = %d, want 1", index.Format)
	}
	if index.Game != Minecraft {
		t.Errorf("Game = %q, want %q", index.Game, Minecraft)
	}
	if index.Name != "Example Modpack" {
		t.Errorf("Name = %q, want %q", index.Name, "Example Modpack")
	}
	if index.Version != "0.1.0-indev" {
		t.Errorf("Version = %q, want %q", index.Version, "0.1.0-indev")
	}
	if index.Deps.Minecraft != "1.20.4" {
		t.Errorf("Deps.Minecraft = %q, want %q", index.Deps.Minecraft, "1.20.4")
	}
	if index.Deps.Fabric != "0.16.10" {
		t.Errorf("Deps.Fabric = %q, want %q", index.Deps.Fabric, "0.16.10")
	}
	if len(index.Files) != 5 {
		t.Fatalf("len(Files) = %d, want 5", len(index.Files))
	}
	f := index.Files[0]
	if f.Path != "mods/menumusic-1.20.4-1.6-fabric.jar" {
		t.Errorf("Files[0].Path = %q", f.Path)
	}
	if f.Hashes.Sha1 != "641841fafaf724308976a809df9e8581ff8aee6a" {
		t.Errorf("Files[0].Hashes.Sha1 = %q", f.Hashes.Sha1)
	}
	if f.Env.Server != modrinth.RequiredEnvSupport {
		t.Errorf("Files[0].Env.Server = %q, want %q", f.Env.Server, modrinth.RequiredEnvSupport)
	}
	if f.FileSize != 14691 {
		t.Errorf("Files[0].FileSize = %d, want 14691", f.FileSize)
	}
}

// Not parallel: ExtractOverrides resolves override paths against the CWD,
// so the test chdirs into the server dir the way cmd does.
func TestExtractOverrides(t *testing.T) {
	pack, err := filepath.Abs(examplePack)
	if err != nil {
		t.Fatal(err)
	}
	serverDir := t.TempDir()
	t.Chdir(serverDir)

	err = ExtractOverrides(pack, serverDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"common-example.txt", "server-example.txt"} {
		data, err := os.ReadFile(filepath.Join(serverDir, name))
		if err != nil {
			t.Errorf("override %s not extracted: %v", name, err)
			continue
		}
		if string(data) != "\n" {
			t.Errorf("%s = %q, want %q", name, data, "\n")
		}
	}
	_, err = os.Stat(filepath.Join(serverDir, "client-example.txt"))
	if !os.IsNotExist(err) {
		t.Errorf("client override extracted, stat err = %v", err)
	}
}

func TestOverrideHashes(t *testing.T) {
	t.Parallel()
	hashes := OverrideHashes(examplePack)
	if len(hashes) != 2 {
		t.Errorf("len(hashes) = %d, want 2: %v", len(hashes), hashes)
	}
	for _, name := range []string{"common-example.txt", "server-example.txt"} {
		h, ok := hashes[name]
		if !ok {
			t.Errorf("no hashes for %s", name)
			continue
		}
		if h.Sha1 != newlineSha1 {
			t.Errorf("%s sha1 = %q, want %q", name, h.Sha1, newlineSha1)
		}
		if h.Sha512 != newlineSha512 {
			t.Errorf("%s sha512 = %q, want %q", name, h.Sha512, newlineSha512)
		}
	}
}

// hashes of the bytes "server"
const serverSha1 = "3de4f901fffb30ac720b0e7eb654b4faa2dd03fa"
const serverSha512 = "fb7daafb37d21221f75d0451fa35dfaacfca509150cecad0e40217b593c5b47566e80c1dd1f74556b3c46a357b699c860976372361a99332934e720a586b7786"

// writeZip stores entries in the given order and returns the archive path.
func writeZip(t *testing.T, entries [][2]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "pack.mrpack")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for _, e := range entries {
		fw, err := w.Create(e[0])
		if err != nil {
			t.Fatal(err)
		}
		_, err = fw.Write([]byte(e[1]))
		if err != nil {
			t.Fatal(err)
		}
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Not parallel: ExtractOverrides needs the CWD set to the server dir.
func TestServerOverridesWin(t *testing.T) {
	server := [2]string{"server-overrides/x", "server"}
	common := [2]string{"overrides/x", "common"}
	orders := []struct {
		name    string
		entries [][2]string
	}{
		{"server first", [][2]string{server, common}},
		{"common first", [][2]string{common, server}},
	}
	for _, order := range orders {
		t.Run(order.name, func(t *testing.T) {
			pack := writeZip(t, order.entries)
			serverDir := t.TempDir()
			t.Chdir(serverDir)

			err := ExtractOverrides(pack, serverDir, nil)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(serverDir, "x"))
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "server" {
				t.Errorf("x = %q, want %q", data, "server")
			}

			h := OverrideHashes(pack)["x"]
			if h.Sha1 != serverSha1 {
				t.Errorf("x sha1 = %q, want %q", h.Sha1, serverSha1)
			}
			if h.Sha512 != serverSha512 {
				t.Errorf("x sha512 = %q, want %q", h.Sha512, serverSha512)
			}
		})
	}
}

// A server file without download URLs stays in the list, so the downloader
// reports it as failed instead of the install silently lacking it.
func TestServerDownloadsKeepsFileWithoutUrls(t *testing.T) {
	t.Parallel()
	index := Index{Files: []File{
		{Path: "mods/nourl.jar", Env: Env{Server: modrinth.RequiredEnvSupport}},
	}}

	dls := index.ServerDownloads(func(File) bool { return true })

	if len(dls) != 1 || dls[0].Path != "mods/nourl.jar" {
		t.Errorf("ServerDownloads() = %v, want mods/nourl.jar", dls)
	}
}

// Not parallel: ExtractOverrides needs the CWD set to the server dir.
func TestExtractOverridesFilter(t *testing.T) {
	pack := writeZip(t, [][2]string{{"overrides/a", "a"}, {"overrides/b", "b"}})
	serverDir := t.TempDir()
	t.Chdir(serverDir)

	err := ExtractOverrides(pack, serverDir, func(p string) bool { return p == "a" })
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(serverDir, "a")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(serverDir, "b")); !os.IsNotExist(err) {
		t.Errorf("b: stat error = %v, want not exist", err)
	}
}

// Not parallel: ExtractOverrides needs the CWD set to the server dir.
func TestExtractOverridesRejectsDuplicates(t *testing.T) {
	pack := writeZip(t, [][2]string{{"overrides/x", "first"}, {"overrides/x", "second"}})
	serverDir := t.TempDir()
	t.Chdir(serverDir)

	err := ExtractOverrides(pack, serverDir, nil)

	if err == nil || !strings.Contains(err.Error(), "overrides/x") {
		t.Errorf("ExtractOverrides() = %v, want an error naming overrides/x", err)
	}
}

// The same path in overrides/ and server-overrides/ is the layering the
// format defines, not a duplicate.
func TestExtractOverridesAcceptsLayeredPath(t *testing.T) {
	pack := writeZip(t, [][2]string{{"overrides/x", "common"}, {"server-overrides/x", "server"}})
	serverDir := t.TempDir()
	t.Chdir(serverDir)

	if err := ExtractOverrides(pack, serverDir, nil); err != nil {
		t.Fatal(err)
	}
}
