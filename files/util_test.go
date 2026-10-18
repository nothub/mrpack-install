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
