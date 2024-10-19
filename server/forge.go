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

type ForgeInstaller struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (inst *ForgeInstaller) Install(serverDir string, serverFile string) error {

	version := inst.ForgeVersion

	if version == "" || version == "latest" {

		meta, err := maven.FetchMetadata("https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml")
		if err != nil {
			return err
		}

		var versions []string
		for _, ver := range meta.Versioning.Versions {
			if !strings.HasPrefix(ver, inst.MinecraftVersion) {
				// does not match mc version
				continue
			}
			versions = append(versions, ver)
		}

		parsed, err := semver.ParseAll(versions)
		if err != nil {
			return err
		}

		sorted := semver.SortDesc(parsed)

		if len(sorted) <= 0 {
			return fmt.Errorf("no forge release version found for mc version %s", inst.MinecraftVersion)
		}

		version = sorted[0].String()
	}

	u := fmt.Sprintf(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar",
		version,
		version,
	)
	installerFile, err := web.DefaultClient.DownloadFile(u, ".", "")
	if err != nil {
		return err
	}

	cmd := exec.Command("java", "-jar", installerFile, "--installServer", serverDir)
	log.Println("Executing command:", cmd.String())
	if err = cmd.Run(); err != nil {
		return err
	}

	if serverFile != "" {
		log.Println("ignoring --server-file option for forge server installation!")
	}

	return nil
}
