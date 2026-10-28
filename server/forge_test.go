package server

import (
	"errors"
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

func TestForgeStartFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		version     string
		listing     []string
		serverFile  string
		want        string
		wantStarter bool
		wantGone    string
	}{
		{"1.12.2 universal", "1.12.2-14.23.5.2768", []string{"forge-1.12.2-14.23.5.2768-universal.jar"}, "mc.jar", "mc.jar", false, "forge-1.12.2-14.23.5.2768-universal.jar"},
		{"1.16.5", "1.16.5-36.2.39", []string{"forge-1.16.5-36.2.39.jar"}, "mc.jar", "mc.jar", false, "forge-1.16.5-36.2.39.jar"},
		{"1.17 to 1.20.2", "1.20.1-47.3.0", []string{"run.sh", "user_jvm_args.txt"}, "mc.jar", "mc.jar", true, ""},
		{"1.20.3 and later", "1.20.4-49.1.0", []string{"run.sh", "forge-1.20.4-49.1.0-shim.jar"}, "mc.jar", "mc.jar", true, ""},
		{"no --server-file", "1.20.4-49.1.0", []string{"run.sh", "forge-1.20.4-49.1.0-shim.jar"}, "", "forge-1.20.4-49.1.0-shim.jar", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, name := range tt.listing {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
					t.Fatal(err)
				}
			}
			starter := false
			placeStarter := func(serverDir string, serverFile string) error {
				starter = true
				return os.WriteFile(filepath.Join(serverDir, serverFile), []byte("starter"), 0644)
			}

			got, err := forgeStartFile(dir, tt.version, tt.serverFile, placeStarter)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.want || starter != tt.wantStarter {
				t.Errorf("forgeStartFile() = %q, starter %v, want %q, starter %v", got, starter, tt.want, tt.wantStarter)
			}
			if tt.serverFile != "" {
				if _, err := os.Stat(filepath.Join(dir, tt.serverFile)); err != nil {
					t.Errorf("%s: %v", tt.serverFile, err)
				}
			}
			if tt.wantGone != "" {
				if _, err := os.Stat(filepath.Join(dir, tt.wantGone)); !os.IsNotExist(err) {
					t.Errorf("%s: stat error = %v, want it renamed", tt.wantGone, err)
				}
			}
		})
	}
}

func TestForgeStartFileStarterFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	fail := func(string, string) error { return errors.New("download failed") }

	if got, err := forgeStartFile(dir, "1.20.1-47.3.0", "mc.jar", fail); err == nil {
		t.Errorf("forgeStartFile() = %q, nil, want an error", got)
	}
}

func TestForgeRejectsVanillaJarName(t *testing.T) {
	t.Parallel()
	inst := ForgeInstaller{MinecraftVersion: "1.16.5", ForgeVersion: "36.2.39"}

	if _, err := inst.Install(t.TempDir(), "minecraft_server.1.16.5.jar"); err == nil {
		t.Error("Install() = nil, want an error")
	}
}
