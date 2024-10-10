package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"log"
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
		// TODO: instead of using the latest release, match against the latest
		//       non-pre-release version that matches the minecraft version
		version = meta.Versioning.Release
	}

	u := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", version, version)
	file, err := web.DefaultClient.DownloadFile(u, serverDir, serverFile)
	if err != nil {
		return err
	}

	log.Println("Server jar downloaded to:", file)
	return nil
}
