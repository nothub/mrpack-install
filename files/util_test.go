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

	// resolve file
	resolved, err := Resolve(file.Name())
	if resolved != file.Name() {
		t.Logf("resolved: %q but should be: %q\n", resolved, file.Name())
		t.FailNow()
	}

	// resolve symlink
	resolved, err = Resolve(link)
	if resolved == link {
		t.Logf("resolved: %q but should be: %q\n", resolved, file.Name())
		t.FailNow()
	}
}
