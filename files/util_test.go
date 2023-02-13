package files

import (
	"testing"
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
