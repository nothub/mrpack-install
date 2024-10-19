package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/semver"
	"log"
	"os/exec"
	"strings"
)

type NeoForgeInstaller struct {
	MinecraftVersion string
	NeoForgeVersion  string
}

func (inst *NeoForgeInstaller) Install(serverDir string, serverFile string) error {
	version := inst.NeoForgeVersion
	if version == "" || version == "latest" {

		meta, err := maven.FetchMetadata("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml")
		if err != nil {
			return err
		}

		mcPrefix := strings.Join(strings.Split(inst.MinecraftVersion, ".")[1:3], ".")
		var versions []string
		for _, ver := range meta.Versioning.Versions {
			if !strings.HasPrefix(ver, mcPrefix) {
				// does not match mc version
				continue
			}
			if strings.Contains(ver, "-") {
				// is pre-release version
				continue
			}
			versions = append(versions, ver)
		}

		parsed, err := semver.ParseAll(versions)
		if err != nil {
			return err
		}

		var releases []semver.Version
		for _, v := range parsed {
			if v.IsRelease() {
				releases = append(releases, v)
			}
		}

		sorted := semver.SortDesc(releases)

		if len(sorted) <= 0 {
			return fmt.Errorf("no neoforge release version found for mc version %s", inst.MinecraftVersion)
		}

		version = sorted[0].String()
	}

	u := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", version, version)
	installerFile, err := web.DefaultClient.DownloadFile(u, serverDir, fmt.Sprintf("neoforge-%s-installer.jar", version))
	if err != nil {
		return err
	}

	log.Println("Installer downloaded to:", installerFile)

	cmd := exec.Command("java", "-jar", installerFile, "--install-server", serverDir)
	log.Println("Executing command:", cmd.String())
	if err = cmd.Run(); err != nil {
		return err
	}

	if serverFile != "" {
		log.Println("ignoring --server-file option for neoforged server installation!")
	}

	return nil
}
