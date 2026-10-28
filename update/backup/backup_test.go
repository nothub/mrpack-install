package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func runDirs(t *testing.T, base string) []string {
	t.Helper()
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		dirs = append(dirs, e.Name())
	}
	return dirs
}

func TestDirMoveKeepsRelativePath(t *testing.T) {
	t.Parallel()
	serverDir := t.TempDir()
	base := filepath.Join(t.TempDir(), "backups")
	writeFile(t, filepath.Join(serverDir, "config", "a.txt"), "mine")

	d := NewDir(base)
	if err := d.Move(serverDir, "config/a.txt"); err != nil {
		t.Fatal(err)
	}

	dirs := runDirs(t, base)
	if len(dirs) != 1 {
		t.Fatalf("backup base holds %v, want one run dir", dirs)
	}
	got, err := os.ReadFile(filepath.Join(base, dirs[0], "config", "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "mine" {
		t.Errorf("backup = %q, want %q", got, "mine")
	}
	if _, err := os.Stat(filepath.Join(serverDir, "config", "a.txt")); !os.IsNotExist(err) {
		t.Errorf("server file: stat error = %v, want not exist", err)
	}
}

func TestNewDirCreatesNothing(t *testing.T) {
	t.Parallel()
	base := filepath.Join(t.TempDir(), "backups")

	NewDir(base)

	if dirs := runDirs(t, base); dirs != nil {
		t.Errorf("backup base holds %v, want nothing", dirs)
	}
}

// Two runs in the same second must not share a run dir, so the second
// can't overwrite the first one's backup.
func TestDirNeverReusesRunDir(t *testing.T) {
	t.Parallel()
	serverDir := t.TempDir()
	base := filepath.Join(t.TempDir(), "backups")

	for _, data := range []string{"first", "second"} {
		writeFile(t, filepath.Join(serverDir, "a.txt"), data)
		if err := NewDir(base).Move(serverDir, "a.txt"); err != nil {
			t.Fatal(err)
		}
	}

	dirs := runDirs(t, base)
	if len(dirs) != 2 {
		t.Fatalf("backup base holds %v, want two run dirs", dirs)
	}
	var got []string
	for _, d := range dirs {
		b, err := os.ReadFile(filepath.Join(base, d, "a.txt"))
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, string(b))
	}
	if !(got[0] == "first" && got[1] == "second" || got[0] == "second" && got[1] == "first") {
		t.Errorf("backups = %v, want first and second", got)
	}
}

func TestDirKeepsOneRunDirPerRun(t *testing.T) {
	t.Parallel()
	serverDir := t.TempDir()
	base := filepath.Join(t.TempDir(), "backups")
	writeFile(t, filepath.Join(serverDir, "a.txt"), "a")
	writeFile(t, filepath.Join(serverDir, "b.txt"), "b")

	d := NewDir(base)
	for _, rel := range []string{"a.txt", "b.txt"} {
		if err := d.Move(serverDir, rel); err != nil {
			t.Fatal(err)
		}
	}

	if dirs := runDirs(t, base); len(dirs) != 1 {
		t.Errorf("backup base holds %v, want one run dir", dirs)
	}
}
