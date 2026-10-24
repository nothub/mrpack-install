package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	_ "unsafe"
)

func testPathIsSubpath(t *testing.T, path string, basePath string) bool {
	ok, err := IsSubpath(path, basePath)
	if err != nil {
		t.Log(err)
	}
	t.Logf("subpath=%-5v path=%s base=%s", ok, path, basePath)
	return ok
}

func TestPathTraversalAbsolute(t *testing.T) {
	if testPathIsSubpath(t, "/bin/file", "/tmp/") {
		t.FailNow()
	}
	if !testPathIsSubpath(t, "/tmp/file", "/tmp/") {
		t.FailNow()
	}
}

func TestPathTraversalRelative(t *testing.T) {
	if testPathIsSubpath(t, "../../../../../../../bin/file", "/tmp/") {
		t.FailNow()
	}
	if !testPathIsSubpath(t, "../../../../../../../tmp/file", "/tmp/") {
		t.FailNow()
	}
}

func TestPathTraversalSiblingPrefix(t *testing.T) {
	base := filepath.Join(t.TempDir(), "mc")
	if testPathIsSubpath(t, filepath.Join(base, "../mc-evil/file"), base) {
		t.FailNow()
	}
	if !testPathIsSubpath(t, filepath.Join(base, "mc-evil/file"), base) {
		t.FailNow()
	}
	if !testPathIsSubpath(t, base, base) {
		t.FailNow()
	}
}

func TestResolveSymlink(t *testing.T) {

	file, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	link := filepath.Join(filepath.Dir(file.Name()), fmt.Sprintf("test-%s", RandString(10)))
	err = os.Symlink(file.Name(), link)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(link)

	// TMPDIR may itself sit behind a symlink (macOS /var -> /private/var)
	want, err := filepath.EvalSymlinks(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// resolve file
	resolved, err := Resolve(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolved: %q but should be: %q", resolved, want)
	}

	// resolve symlink
	resolved, err = Resolve(link)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolved: %q but should be: %q", resolved, want)
	}
}

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
