package server

import (
	"errors"
	"fmt"
	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/http"
	"os"
	"os/exec"
	"path"
)

const quiltInstallerUrl = "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/0.5.0/quilt-installer-0.5.0.jar"

type QuiltInstaller struct {
	MinecraftVersion string
	QuiltVersion     string
}

func (inst *QuiltInstaller) Install(serverDir string, serverFile string) error {
	installer, err := http.DefaultClient.DownloadFile(quiltInstallerUrl, ".", "")
	if err != nil {
		return err
	}

	cmd := exec.Command("java", "-jar", installer, "install", "server", inst.MinecraftVersion, "--install-dir="+serverDir, "--create-scripts", "--download-server")
	fmt.Println("Executing command:", cmd.String())
	err = cmd.Run()
	if err != nil {
		return err
	}

	if !files.IsFile(path.Join(serverDir, "server.jar")) {
		return errors.New("server.jar not found")
	}

	if serverFile != "" {
		err = os.Rename(path.Join(serverDir, "server.jar"), path.Join(serverDir, serverFile))
		if err != nil {
			return err
		}
	}

	return nil
}
