package mrpack

import (
	"os"
	"path/filepath"
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

	err = ExtractOverrides(pack, serverDir)
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
