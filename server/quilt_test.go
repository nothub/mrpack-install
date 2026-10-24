package server

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestQuiltInstallerArgs(t *testing.T) {
	t.Parallel()
	tail := []string{"--install-dir=/srv/mc", "--create-scripts", "--download-server"}
	tests := map[string][]string{
		"0.20.1": append([]string{"-jar", "installer.jar", "install", "server", "1.20.1", "0.20.1"}, tail...),
		"":       append([]string{"-jar", "installer.jar", "install", "server", "1.20.1"}, tail...),
		"latest": append([]string{"-jar", "installer.jar", "install", "server", "1.20.1"}, tail...),
	}
	for version, want := range tests {
		inst := QuiltInstaller{MinecraftVersion: "1.20.1", QuiltVersion: version}

		got := inst.args("installer.jar", "/srv/mc")

		if !reflect.DeepEqual(got, want) {
			t.Errorf("version %q: args = %v, want %v", version, got, want)
		}
	}
}

// writeLaunchJar writes a jar whose manifest has the given Class-Path,
// folded into lines of at most 72 bytes, continuations included.
func writeLaunchJar(t *testing.T, classPath string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "quilt-server-launch.jar")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("META-INF/MANIFEST.MF")
	if err != nil {
		t.Fatal(err)
	}
	line := "Class-Path: " + classPath
	var b strings.Builder
	b.WriteString("Manifest-Version: 1.0\r\n")
	width := 72
	for len(line) > width {
		b.WriteString(line[:width] + "\r\n ")
		line = line[width:]
		width = 71
	}
	b.WriteString(line + "\r\nMain-Class: org.quiltmc.loader.impl.launch.server.QuiltServerLauncher\r\n\r\n")
	if _, err := w.Write([]byte(b.String())); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

const loaderClassPath = "libraries/org/ow2/asm/asm/9.6/asm-9.6.jar libraries/org/quiltmc/quilt-loader/0.20.1/quilt-loader-0.20.1.jar libraries/net/fabricmc/intermediary/1.20.1/intermediary-1.20.1.jar"

func TestCheckLaunchJar(t *testing.T) {
	t.Parallel()
	jar := writeLaunchJar(t, loaderClassPath)

	for _, version := range []string{"0.20.1", "", "latest"} {
		if err := checkLaunchJar(jar, version); err != nil {
			t.Errorf("version %q: %v", version, err)
		}
	}
	err := checkLaunchJar(jar, "0.21.0")
	if err == nil || !strings.Contains(err.Error(), "quilt-loader-0.20.1.jar") {
		t.Errorf("version 0.21.0: %v, want an error naming the loaded quilt-loader-0.20.1.jar", err)
	}
	if err := checkLaunchJar(jar, "0.20.*"); err == nil {
		t.Error("version 0.20.*: nil, want the version compared literally")
	}
}

func TestCheckLaunchJarLowercaseAttribute(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "quilt-server-launch.jar")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("META-INF/MANIFEST.MF")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("Manifest-Version: 1.0\r\nclass-path: quilt-loader-0.20.1.jar\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := checkLaunchJar(path, "0.20.1"); err != nil {
		t.Error(err)
	}
}

func TestCheckLaunchJarWithoutLoader(t *testing.T) {
	t.Parallel()
	jar := writeLaunchJar(t, "libraries/org/ow2/asm/asm/9.6/asm-9.6.jar")

	if err := checkLaunchJar(jar, ""); err == nil {
		t.Error("checkLaunchJar() = nil, want an error")
	}
}

func TestCheckLaunchJarMissing(t *testing.T) {
	t.Parallel()

	if err := checkLaunchJar(filepath.Join(t.TempDir(), "quilt-server-launch.jar"), ""); err == nil {
		t.Error("checkLaunchJar() = nil, want an error")
	}
}

func TestQuiltInstallRejectsVanillaJarName(t *testing.T) {
	t.Parallel()
	inst := QuiltInstaller{MinecraftVersion: "1.20.1", QuiltVersion: "0.20.1"}

	if _, err := inst.Install(t.TempDir(), "server.jar"); err == nil {
		t.Error("Install() = nil, want an error")
	}
}

func TestQuiltInstallRejectsFlagLikeVersion(t *testing.T) {
	t.Parallel()
	inst := QuiltInstaller{MinecraftVersion: "1.20.1", QuiltVersion: "--install-dir=/tmp"}

	if _, err := inst.Install(t.TempDir(), ""); err == nil {
		t.Error("Install() = nil, want an error")
	}
}
