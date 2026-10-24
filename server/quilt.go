package server

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type QuiltInstaller struct {
	MinecraftVersion string
	QuiltVersion     string
}

// quiltLaunchJar is the jar the Quilt installer writes to start the
// server. It loads the vanilla server.jar next to it.
const quiltLaunchJar = "quilt-server-launch.jar"

func (inst *QuiltInstaller) Install(serverDir string, serverFile string) error {
	if serverFile == "server.jar" {
		return errors.New("--server-file server.jar would replace the vanilla jar the Quilt launcher loads, choose another name")
	}
	if strings.HasPrefix(inst.QuiltVersion, "-") {
		return fmt.Errorf("invalid Quilt loader version %q", inst.QuiltVersion)
	}

	meta, err := maven.FetchMetadata("https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/maven-metadata.xml")
	if err != nil {
		return err
	}
	quiltInstallerUrl := "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/" + meta.Versioning.Release + "/quilt-installer-" + meta.Versioning.Release + ".jar"

	installer, err := web.DefaultClient.DownloadFile(quiltInstallerUrl, ".", "")
	if err != nil {
		return err
	}

	// A launch jar from an earlier install would pass the check below.
	err = os.Remove(filepath.Join(serverDir, quiltLaunchJar))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	cmd := exec.Command("java", inst.args(installer, serverDir)...)
	log.Println("Executing command:", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}

	// The Quilt installer exits 0 on every failure, so the jar it writes is
	// the only proof the install worked, and its output the only reason.
	err = checkLaunchJar(filepath.Join(serverDir, quiltLaunchJar), inst.QuiltVersion)
	if err != nil {
		return fmt.Errorf("quilt install failed: %w\n%s", err, out)
	}

	if serverFile != "" {
		err = os.Rename(filepath.Join(serverDir, quiltLaunchJar), filepath.Join(serverDir, serverFile))
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *QuiltInstaller) args(installer string, serverDir string) []string {
	args := []string{"-jar", installer, "install", "server", inst.MinecraftVersion}
	if quiltPinned(inst.QuiltVersion) {
		args = append(args, inst.QuiltVersion)
	}
	return append(args, "--install-dir="+serverDir, "--create-scripts", "--download-server")
}

func quiltPinned(version string) bool {
	return version != "" && version != "latest"
}

// checkLaunchJar fails unless the launch jar at jarPath exists and its
// manifest Class-Path names quilt-loader-<version>.jar, or any loader jar
// when version is not pinned.
func checkLaunchJar(jarPath string, version string) error {
	name := filepath.Base(jarPath)
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return err
	}
	defer r.Close()
	f, err := r.Open("META-INF/MANIFEST.MF")
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	defer f.Close()

	// Manifest lines are folded at 72 bytes, a continuation starts with
	// one space: https://docs.oracle.com/en/java/javase/17/docs/specs/jar/jar.html
	var unfolded []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.HasPrefix(line, " ") && len(unfolded) > 0 {
			unfolded[len(unfolded)-1] += line[1:]
			continue
		}
		unfolded = append(unfolded, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	want := "quilt-loader-" + version + ".jar"
	var loaders []string
	for _, line := range unfolded {
		attr, value, ok := strings.Cut(line, ": ")
		if !ok || !strings.EqualFold(attr, "Class-Path") {
			continue
		}
		for _, entry := range strings.Fields(value) {
			base := path.Base(entry)
			if !strings.HasPrefix(base, "quilt-loader-") || !strings.HasSuffix(base, ".jar") {
				continue
			}
			if !quiltPinned(version) || base == want {
				return nil
			}
			loaders = append(loaders, base)
		}
	}
	if !quiltPinned(version) {
		return fmt.Errorf("%s loads no quilt-loader jar", name)
	}
	return fmt.Errorf("%s loads %v, want %s", name, loaders, want)
}
