package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/semver"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NeoForgeInstaller struct {
	MinecraftVersion string
	NeoForgeVersion  string
}

func (inst *NeoForgeInstaller) Install(serverDir string, serverFile string) (string, error) {
	if !neoforgeSupports(inst.MinecraftVersion) {
		return "", fmt.Errorf("NeoForge starts at Minecraft 1.20.1, there is none for %s", inst.MinecraftVersion)
	}
	version := inst.NeoForgeVersion
	if version == "" || version == "latest" {

		metaUrl := "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml"
		if inst.MinecraftVersion == "1.20.1" {
			metaUrl = "https://maven.neoforged.net/releases/net/neoforged/forge/maven-metadata.xml"
		}
		meta, err := maven.FetchMetadata(metaUrl)
		if err != nil {
			return "", err
		}

		parsed, err := semver.ParseAll(neoforgeVersionsFor(meta.Versioning.Versions, inst.MinecraftVersion))
		if err != nil {
			return "", err
		}

		var releases []semver.Version
		for _, v := range parsed {
			if v.IsRelease() {
				releases = append(releases, v)
			}
		}

		sorted := semver.SortDesc(releases)

		if len(sorted) <= 0 {
			return "", fmt.Errorf("no neoforge release version found for mc version %s", inst.MinecraftVersion)
		}

		version = sorted[0].String()
	}

	u := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", version, version)
	if inst.MinecraftVersion == "1.20.1" {
		u = fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/forge/1.20.1-%s/forge-1.20.1-%s-installer.jar", version, version)
	}
	installerFile, err := web.DefaultClient.DownloadFile(u, serverDir, fmt.Sprintf("neoforge-%s-installer.jar", version))
	if err != nil {
		return "", err
	}

	log.Println("Installer downloaded to:", installerFile)

	args := []string{"-jar", installerFile, "--install-server", serverDir}
	if serverFile != "" {
		// The installer then also writes ServerStarterJar as server.jar,
		// which starts NeoForge with java -jar.
		args = append(args, "--server-jar")
	}
	cmd := exec.Command("java", args...)
	log.Println("Executing command:", cmd.String())
	if err = cmd.Run(); err != nil {
		return "", err
	}

	if serverFile != "" && serverFile != "server.jar" {
		err = os.Rename(filepath.Join(serverDir, "server.jar"), filepath.Join(serverDir, serverFile))
		if err != nil {
			return "", fmt.Errorf("renaming the NeoForge server.jar: %w", err)
		}
	}
	removeInstaller(installerFile)
	// serverFile is empty when only run.sh and run.bat start NeoForge.
	return serverFile, nil
}

// neoforgeVersionsFor returns the versions without a pre-release suffix
// that belong to Minecraft version mc. NeoForge numbers them
// <mc minor>.<mc patch>.<build>, so 21.1.x is for 1.21.1 and 21.11.x for
// 1.21.11, and a two-part mc version has patch 0. For 1.20.1 NeoForge is
// the fork of Forge, numbered 1.20.1-<forge version>, and the prefix is
// cut because the download URL adds it back.
func neoforgeVersionsFor(versions []string, mc string) []string {
	if mc == "1.20.1" {
		var matching []string
		for _, v := range versions {
			if forge, ok := strings.CutPrefix(v, "1.20.1-"); ok && !strings.Contains(forge, "-") {
				matching = append(matching, forge)
			}
		}
		return matching
	}
	parts := strings.Split(mc, ".")
	if len(parts) < 2 {
		return nil
	}
	patch := "0"
	if len(parts) > 2 {
		patch = parts[2]
	}
	prefix := parts[1] + "." + patch + "."
	var matching []string
	for _, v := range versions {
		if strings.HasPrefix(v, prefix) && !strings.Contains(v, "-") {
			matching = append(matching, v)
		}
	}
	return matching
}

// neoforgeSupports reports whether NeoForge exists for Minecraft version
// mc: from 1.20.1, where it forked from Forge, on. A version it can't read
// is left for the version lookup to reject.
func neoforgeSupports(mc string) bool {
	var major, minor, patch int
	n, _ := fmt.Sscanf(mc, "%d.%d.%d", &major, &minor, &patch)
	if n < 2 {
		return true
	}
	switch {
	case major != 1:
		return major > 1
	case minor != 20:
		return minor > 20
	default:
		return patch >= 1
	}
}
