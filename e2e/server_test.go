package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServer(t *testing.T) {
	t.Parallel()
	for _, flavor := range []string{"vanilla", "fabric", "paper"} {
		t.Run(flavor, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			stdout, out, err := mrpackInstallStdout(t, "server", flavor, "--minecraft-version", "1.21.1", "--server-dir", dir)
			if err != nil {
				t.Fatalf("mrpack-install server %s: %v\n%s", flavor, err, out)
			}

			// Without --server-file, each flavor names the jar after its download.
			jars, err := filepath.Glob(filepath.Join(dir, "*.jar"))
			if err != nil {
				t.Fatal(err)
			}
			if len(jars) != 1 {
				t.Fatalf("server jars = %v, want one\n%s", jars, out)
			}
			fi, err := os.Stat(jars[0])
			if err != nil {
				t.Fatal(err)
			}
			if fi.Size() == 0 {
				t.Errorf("%s is empty", jars[0])
			}
			// Logs go to stderr, so a script gets only the result.
			if got, want := string(stdout), filepath.Base(jars[0])+"\n"; got != want {
				t.Errorf("stdout = %q, want %q", got, want)
			}
		})
	}
}
