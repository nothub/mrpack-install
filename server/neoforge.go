package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/web"
	"log"
)

type NeoForgeInstaller struct {
	MinecraftVersion string
	NeoForgeVersion  string
}

// TODO maven version lookup: https://maven.neoforged.net/releases/net/neoforged/forge/maven-metadata.xml

func (inst *NeoForgeInstaller) Install(serverDir string, serverFile string) error {
	// TODO: implement automatic lookup for latest version
	if inst.NeoForgeVersion == "" || inst.NeoForgeVersion == "latest" {
		log.Fatalln("automatic NeoForge version lookup not implemented\nplease set server version with --flavor-version flag")
	}
	u := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", inst.NeoForgeVersion, inst.NeoForgeVersion)
	file, err := web.DefaultClient.DownloadFile(u, serverDir, serverFile)
	if err != nil {
		return err
	}
	log.Println("Server jar downloaded to:", file)
	return nil
}
