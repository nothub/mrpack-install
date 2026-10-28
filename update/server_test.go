package update

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
)

// fakeInstaller writes file with content into the server dir, like an
// installer that produced that server jar. An empty file is a server that
// only starts through run.sh.
type fakeInstaller struct {
	file       string
	content    string
	err        error
	reportOnly bool
	calls      int
}

func (f *fakeInstaller) Install(serverDir string, serverFile string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	if f.file == "" {
		return "", os.WriteFile(filepath.Join(serverDir, "run.sh"), []byte(f.content), 0755)
	}
	file := f.file
	if serverFile != "" {
		file = serverFile
	}
	if f.reportOnly {
		return file, nil
	}
	return file, os.WriteFile(filepath.Join(serverDir, file), []byte(f.content), 0644)
}

func withServerFile(path string, content string) *packstate.Schema {
	s := state(map[string]modrinth.Hashes{})
	s.ServerFile = &packstate.ServerFile{Path: path, Hashes: hashes(content)}
	return s
}

func newState() *packstate.Schema {
	return state(map[string]modrinth.Hashes{})
}

func TestInstallServerRecordsServerFile(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, nil)
	inst := &fakeInstaller{file: "server-1.jar", content: "v1"}
	st := newState()

	err := InstallServer(dir, &Plan{InstallServer: true}, nil, st, inst, "", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	want := packstate.ServerFile{Path: "server-1.jar", Hashes: hashes("v1")}
	if st.ServerFile == nil || *st.ServerFile != want {
		t.Errorf("ServerFile = %+v, want %+v", st.ServerFile, want)
	}
}

func TestInstallServerKeepsRecordWithoutReinstall(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"server-1.jar": "v1"})
	inst := &fakeInstaller{file: "server-2.jar", content: "v2"}
	old := withServerFile("server-1.jar", "v1")
	st := newState()

	err := InstallServer(dir, &Plan{}, old, st, inst, "", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	if inst.calls != 0 {
		t.Errorf("installer calls = %d, want 0", inst.calls)
	}
	if st.ServerFile == nil || *st.ServerFile != *old.ServerFile {
		t.Errorf("ServerFile = %+v, want the old record %+v", st.ServerFile, old.ServerFile)
	}
}

func TestInstallServerDeletesRenamedServerFile(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"server-1.jar": "v1"})
	inst := &fakeInstaller{file: "server-2.jar", content: "v2"}
	st := newState()

	err := InstallServer(dir, &Plan{InstallServer: true}, withServerFile("server-1.jar", "v1"), st, inst, "", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	assertNoFile(t, filepath.Join(dir, "server-1.jar"))
	assertNoFile(t, filepath.Join(dir, "backups"))
	assertFile(t, filepath.Join(dir, "server-2.jar"), "v2")
}

func TestInstallServerDeletesServerFileForRunScript(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"forge-1.16.5-36.2.39.jar": "v1"})
	inst := &fakeInstaller{content: "#!/bin/sh"}
	st := newState()

	err := InstallServer(dir, &Plan{InstallServer: true}, withServerFile("forge-1.16.5-36.2.39.jar", "v1"), st, inst, "", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	assertNoFile(t, filepath.Join(dir, "forge-1.16.5-36.2.39.jar"))
	if st.ServerFile == nil || *st.ServerFile != (packstate.ServerFile{}) {
		t.Errorf("ServerFile = %+v, want an empty record for run.sh", st.ServerFile)
	}
}

func TestInstallServerBacksUpModifiedServerFile(t *testing.T) {
	t.Parallel()
	for name, newFile := range map[string]string{"renamed": "server-2.jar", "same name": "server.jar"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := serverDir(t, map[string]string{"server.jar": "patched"})
			base := filepath.Join(dir, "backups")
			inst := &fakeInstaller{file: newFile, content: "v2"}

			err := InstallServer(dir, &Plan{InstallServer: true}, withServerFile("server.jar", "v1"), newState(), inst, "", backup.NewDir(base))
			if err != nil {
				t.Fatal(err)
			}

			runs, err := filepath.Glob(filepath.Join(base, "*", "server.jar"))
			if err != nil {
				t.Fatal(err)
			}
			if len(runs) != 1 {
				t.Fatalf("backed up server jars = %v, want one", runs)
			}
			assertFile(t, runs[0], "patched")
			assertFile(t, filepath.Join(dir, newFile), "v2")
		})
	}
}

func TestInstallServerReplacesUnmodifiedServerFileOfSameName(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"mc.jar": "v1"})
	inst := &fakeInstaller{file: "fabric-server.jar", content: "v2"}

	err := InstallServer(dir, &Plan{InstallServer: true}, withServerFile("mc.jar", "v1"), newState(), inst, "mc.jar", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(dir, "mc.jar"), "v2")
	assertNoFile(t, filepath.Join(dir, "backups"))
}

// A jar at the --server-file name that no run recorded is the admin's, so
// it is backed up before the installer overwrites it.
func TestInstallServerBacksUpUnrecordedServerFile(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"mc.jar": "mine"})
	base := filepath.Join(dir, "backups")
	inst := &fakeInstaller{file: "fabric-server.jar", content: "v1"}

	err := InstallServer(dir, &Plan{InstallServer: true}, nil, newState(), inst, "mc.jar", backup.NewDir(base))
	if err != nil {
		t.Fatal(err)
	}

	runs, err := filepath.Glob(filepath.Join(base, "*", "mc.jar"))
	if err != nil || len(runs) != 1 {
		t.Fatalf("backed up jars = %v, %v, want one", runs, err)
	}
	assertFile(t, runs[0], "mine")
	assertFile(t, filepath.Join(dir, "mc.jar"), "v1")
}

func TestInstallServerFailures(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		old  *packstate.Schema
		inst *fakeInstaller
	}{
		"unsafe recorded path":  {withServerFile("../server.jar", "v1"), &fakeInstaller{file: "server.jar", content: "v2"}},
		"installer fails":       {nil, &fakeInstaller{err: errors.New("java not found")}},
		"reported file missing": {nil, &fakeInstaller{file: "server.jar", reportOnly: true}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := serverDir(t, nil)

			err := InstallServer(dir, &Plan{InstallServer: true}, tt.old, newState(), tt.inst, "", backup.NewDir(filepath.Join(dir, "backups")))

			if err == nil {
				t.Error("InstallServer() = nil, want an error")
			}
		})
	}
}

func assertFile(t *testing.T, path string, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s: stat error = %v, want not exist", path, err)
	}
}

// rewritingInstaller rewrites keep, as the Forge 1.20.3+ installer rewrites
// its shim jar, and reports file as the server file.
type rewritingInstaller struct {
	keep string
	file string
}

func (r *rewritingInstaller) Install(serverDir string, _ string) (string, error) {
	if err := os.WriteFile(filepath.Join(serverDir, r.keep), []byte("shim"), 0644); err != nil {
		return "", err
	}
	return r.file, os.WriteFile(filepath.Join(serverDir, r.file), []byte("starter"), 0644)
}

// A recorded file the install wrote again is part of the new install, so
// it stays even when the install reports another server file.
func TestInstallServerKeepsRecordedFileTheInstallRewrote(t *testing.T) {
	t.Parallel()
	dir := serverDir(t, map[string]string{"forge-shim.jar": "shim"})
	old := withServerFile("forge-shim.jar", "shim")
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(dir, "forge-shim.jar"), past, past); err != nil {
		t.Fatal(err)
	}
	inst := &rewritingInstaller{keep: "forge-shim.jar", file: "mc.jar"}

	err := InstallServer(dir, &Plan{InstallServer: true}, old, newState(), inst, "mc.jar", backup.NewDir(filepath.Join(dir, "backups")))
	if err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(dir, "forge-shim.jar"), "shim")
	assertFile(t, filepath.Join(dir, "mc.jar"), "starter")
}
