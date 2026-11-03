package files

import (
	"os"
	"path/filepath"
	"testing"
	_ "unsafe"
)

func TestInDir(t *testing.T) {
	base := filepath.Join(t.TempDir(), "mc")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(base, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "file"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "missing"), filepath.Join(base, "dangling")); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"mods/a.jar", "config/../mods/a.jar", "mc-evil/file"} {
		got, err := InDir(base, rel)
		if err != nil {
			t.Errorf("InDir(%q) error = %v, want nil", rel, err)
		} else if want := filepath.Join(base, rel); got != want {
			t.Errorf("InDir(%q) = %q, want %q", rel, got, want)
		}
	}
	for _, rel := range []string{"../mc-evil/file", "..", ".", "", "/etc/passwd", "link/file", "link/new/file", "link", "dangling", "dangling/file"} {
		if got, err := InDir(base, rel); err == nil {
			t.Errorf("InDir(%q) = %q, want an error", rel, got)
		}
	}
}

// A server dir reached through a symlink still accepts files that don't
// exist yet.
func TestInDirSymlinkedDir(t *testing.T) {
	real := t.TempDir()
	dir := filepath.Join(t.TempDir(), "mc")
	if err := os.Symlink(real, dir); err != nil {
		t.Fatal(err)
	}

	got, err := InDir(dir, "mods/new.jar")

	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "mods", "new.jar"); got != want {
		t.Errorf("InDir() = %q, want %q", got, want)
	}
}

func TestMove(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "a", "b", "dst")
	if err := os.WriteFile(src, []byte("data"), 0640); err != nil {
		t.Fatal(err)
	}

	if err := Move(src, dst); err != nil {
		t.Fatal(err)
	}

	assertMoved(t, src, dst, "data", 0640)
}

// Not parallel: it swaps the package's rename.
func TestMoveAcrossFilesystems(t *testing.T) {
	rename = func(src string, dst string) error {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: errCrossDevice}
	}
	t.Cleanup(func() { rename = os.Rename })
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "a", "dst")
	if err := os.WriteFile(src, []byte("data"), 0640); err != nil {
		t.Fatal(err)
	}

	if err := Move(src, dst); err != nil {
		t.Fatal(err)
	}

	assertMoved(t, src, dst, "data", 0640)
}

func TestMoveRefusesExistingDst(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	for _, p := range []string{src, dst} {
		if err := os.WriteFile(p, []byte(p), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Move(src, dst); err == nil {
		t.Fatal("Move() = nil, want an error")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != dst {
		t.Errorf("dst = %q, want it unchanged", got)
	}
}

func assertMoved(t *testing.T, src string, dst string, data string, mode os.FileMode) {
	t.Helper()
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src: stat error = %v, want not exist", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != data {
		t.Errorf("dst = %q, want %q", got, data)
	}
	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != mode {
		t.Errorf("dst mode = %v, want %v", fi.Mode().Perm(), mode)
	}
}

// A stat error other than not-exist is returned, not fatal.
func TestIsFileReportsStatErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := IsFile(filepath.Join(file, "child"))

	if err == nil {
		t.Errorf("IsFile(file/child) = %v, nil, want an error", ok)
	}
}

func TestIsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0644); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]bool{file: true, dir: false, filepath.Join(dir, "missing"): false} {
		got, err := IsFile(path)
		if err != nil || got != want {
			t.Errorf("IsFile(%s) = %v, %v, want %v, nil", path, got, err, want)
		}
	}
}
