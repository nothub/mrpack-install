package server

import (
	"errors"
	"github.com/nothub/gorinth/http"
	"net/url"
)

type Fabric struct {
	MinecraftVersion string
	FabricVersion    string
}

func (supplier *Fabric) GetUrl() (string, error) {
	loaderVersion := supplier.FabricVersion
	if loaderVersion == "" || loaderVersion == "latest" {
		var loaders []struct {
			Infos struct {
				Version string `json:"version"`
				Stable  bool   `json:"stable"`
			} `json:"loader"`
		}
		err := http.Instance.GetJson("https://meta.fabricmc.net/v2/versions/loader/"+supplier.MinecraftVersion, nil, &loaders, nil)
		if err != nil {
			return "", err
		}
		for i := range loaders {
			if loaders[i].Infos.Stable {
				loaderVersion = loaders[i].Infos.Version
				break
			}
		}
	}
	if loaderVersion == "" {
		return "", errors.New("no stable fabric loader release found")
	}

	var installerVersion string
	var installers []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	err := http.Instance.GetJson("https://meta.fabricmc.net/v2/versions/installer", nil, &installers, nil)
	if err != nil {
		return "", err
	}
	for i := range installers {
		if installers[i].Stable {
			installerVersion = installers[i].Version
			break
		}
	}
	if loaderVersion == "" {
		return "", errors.New("no stable fabric installer release found")
	}

	versionTriple := supplier.MinecraftVersion + "/" + loaderVersion + "/" + installerVersion
	u, err := url.Parse("https://meta.fabricmc.net/v2/versions/loader/" + versionTriple + "/server/jar")
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
