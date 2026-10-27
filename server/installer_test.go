package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveInstaller(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jar := filepath.Join(dir, "forge-1.20.1-47.3.0-installer.jar")
	for _, p := range []string{jar, jar + ".log", filepath.Join(dir, "run.sh")} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	removeInstaller(jar)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "run.sh" {
		t.Errorf("server dir holds %v, want only run.sh", entries)
	}
}

// Quilt's installer writes no log.
func TestRemoveInstallerWithoutLog(t *testing.T) {
	t.Parallel()
	jar := filepath.Join(t.TempDir(), "quilt-installer-0.15.1.jar")
	if err := os.WriteFile(jar, nil, 0644); err != nil {
		t.Fatal(err)
	}

	removeInstaller(jar)

	if _, err := os.Stat(jar); !os.IsNotExist(err) {
		t.Errorf("%s: stat error = %v, want removed", jar, err)
	}
}

// A log that can't be removed doesn't stop the jar from going.
func TestRemoveInstallerKeepsGoing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	jar := filepath.Join(dir, "forge-1.20.1-47.3.0-installer.jar")
	if err := os.WriteFile(jar, nil, 0644); err != nil {
		t.Fatal(err)
	}
	// A non-empty dir at the log path makes its removal fail.
	if err := os.MkdirAll(filepath.Join(jar+".log", "x"), 0755); err != nil {
		t.Fatal(err)
	}

	removeInstaller(jar)

	if _, err := os.Stat(jar); !os.IsNotExist(err) {
		t.Errorf("%s: stat error = %v, want removed", jar, err)
	}
}
