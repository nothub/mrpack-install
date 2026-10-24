package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/web"
	"log"
	"net/url"
	"path/filepath"
)

type PaperInstaller struct {
	MinecraftVersion string
}

func (inst *PaperInstaller) Install(serverDir string, serverFile string) (string, error) {
	// builds/latest ignores the channel, so ask for the STABLE list. It is ordered newest first.
	var builds []struct {
		Downloads map[string]struct {
			Url string `json:"url"`
		} `json:"downloads"`
	}
	err := web.DefaultClient.GetJson("https://fill.papermc.io/v3/projects/paper/versions/"+
		url.PathEscape(inst.MinecraftVersion)+"/builds?channel=STABLE", &builds, nil)
	if err != nil {
		return "", fmt.Errorf("fetching paper builds for %s: %w", inst.MinecraftVersion, err)
	}
	if len(builds) == 0 {
		return "", fmt.Errorf("no stable paper build for %s", inst.MinecraftVersion)
	}

	download, ok := builds[0].Downloads["server:default"]
	if !ok {
		return "", fmt.Errorf("newest stable paper build for %s has no server:default download", inst.MinecraftVersion)
	}

	file, err := web.DefaultClient.DownloadFile(download.Url, serverDir, serverFile)
	if err != nil {
		return "", fmt.Errorf("downloading paper server jar for %s: %w", inst.MinecraftVersion, err)
	}

	log.Println("Server jar downloaded to:", file)
	return filepath.Base(file), nil
}
