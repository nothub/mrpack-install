package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/web"
	"log"
	"os/exec"
)

type ForgeInstaller struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (inst *ForgeInstaller) Install(serverDir string, serverFile string) error {
	u := fmt.Sprintf(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/forge-%s-%s-installer.jar",
		inst.MinecraftVersion,
		inst.ForgeVersion,
		inst.MinecraftVersion,
		inst.ForgeVersion,
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
