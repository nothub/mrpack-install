package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestForgeServerFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
		listing []string
		want    string
	}{
		{"1.12.2 newer builds", "1.12.2-14.23.5.2859",
			[]string{"forge-1.12.2-14.23.5.2859.jar", "forge-1.12.2-14.23.5.2859-installer.jar", "minecraft_server.1.12.2.jar"},
			"forge-1.12.2-14.23.5.2859.jar"},
		{"1.12.2 older builds", "1.12.2-14.23.5.2768",
			[]string{"forge-1.12.2-14.23.5.2768-universal.jar", "forge-1.12.2-14.23.5.2768-installer.jar", "minecraft_server.1.12.2.jar"},
			"forge-1.12.2-14.23.5.2768-universal.jar"},
		{"1.16.5", "1.16.5-36.2.39",
			[]string{"forge-1.16.5-36.2.39.jar", "forge-1.16.5-36.2.39-installer.jar", "minecraft_server.1.16.5.jar"},
			"forge-1.16.5-36.2.39.jar"},
		{"1.17 to 1.20.2", "1.20.1-47.3.0",
			[]string{"run.sh", "run.bat", "user_jvm_args.txt", "forge-1.20.1-47.3.0-installer.jar"},
			""},
		{"1.20.3 and later", "1.20.4-49.1.0",
			[]string{"run.sh", "run.bat", "user_jvm_args.txt", "forge-1.20.4-49.1.0-shim.jar", "forge-1.20.4-49.1.0-installer.jar"},
			"forge-1.20.4-49.1.0-shim.jar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, name := range tt.listing {
				if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := forgeServerFile(dir, tt.version)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.want {
				t.Errorf("forgeServerFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestForgeServerFileNothingInstalled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "forge-1.20.1-47.3.0-installer.jar"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	if got, err := forgeServerFile(dir, "1.20.1-47.3.0"); err == nil {
		t.Errorf("forgeServerFile() = %q, nil, want an error", got)
	}
}
