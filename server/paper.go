package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/web"
	"log"
	"net/url"
	"path"
	"path/filepath"
)

type PaperInstaller struct {
	MinecraftVersion string
}

func (inst *PaperInstaller) Install(serverDir string, serverFile string) (string, error) {
	// builds/latest ignores the channel, so ask for the STABLE list. It is ordered newest first.
	var builds []struct {
		Downloads map[string]struct {
			Name      string `json:"name"`
			Url       string `json:"url"`
			Checksums struct {
				Sha256 string `json:"sha256"`
			} `json:"checksums"`
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

	name := serverFile
	if name == "" {
		// The name comes from the API, so only its last element counts.
		name = path.Base(download.Name)
		if download.Name == "" {
			name = path.Base(download.Url)
		}
	}
	err = downloadVerified(download.Url, download.Checksums.Sha256, serverDir, name)
	if err != nil {
		return "", fmt.Errorf("downloading paper server jar for %s: %w", inst.MinecraftVersion, err)
	}

	log.Println("Server jar downloaded to:", filepath.Join(serverDir, name))
	return name, nil
}
