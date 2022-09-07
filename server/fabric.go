package server

import (
	"errors"
	"fmt"
	"github.com/nothub/mrpack-install/requester"
	"net/url"
)

type Fabric struct {
	MinecraftVersion string
	FabricVersion    string
}

func (supplier *Fabric) Provide(serverDir string, serverFile string) error {
	loaderVersion := supplier.FabricVersion
	if loaderVersion == "" || loaderVersion == "latest" {
		latestLoaderVersion, err := latestFabricLoaderVersion(supplier.MinecraftVersion)
		if err != nil {
			return err
		}
		loaderVersion = latestLoaderVersion
	}

	installerVersion, err := latestFabricInstallerVersion()
	if err != nil {
		return err
	}

	versionTriple := supplier.MinecraftVersion + "/" + loaderVersion + "/" + installerVersion
	u, err := url.Parse("https://meta.fabricmc.net/v2/versions/loader/" + versionTriple + "/server/jar")
	if err != nil {
		return err
	}

	file, err := requester.DefaultHttpClient.DownloadFile(u.String(), serverDir, serverFile)
	if err != nil {
		return err
	}

	fmt.Println("Server jar downloaded to:", file)
	return nil
}

func latestFabricLoaderVersion(mcVer string) (string, error) {
	var loaders []struct {
		Infos struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		} `json:"loader"`
	}
	err := requester.DefaultHttpClient.GetJson("https://meta.fabricmc.net/v2/versions/loader/"+mcVer, &loaders, nil)
	if err != nil {
		return "", err
	}
	for i := range loaders {
		if loaders[i].Infos.Stable {
			return loaders[i].Infos.Version, nil
		}
	}
	return "", errors.New("no stable fabric loader release found")
}

func latestFabricInstallerVersion() (string, error) {
	var installers []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	err := requester.DefaultHttpClient.GetJson("https://meta.fabricmc.net/v2/versions/installer", &installers, nil)
	if err != nil {
		return "", err
	}
	for i := range installers {
		if installers[i].Stable {
			return installers[i].Version, nil
		}
	}
	return "", errors.New("no stable fabric installer release found")
}
