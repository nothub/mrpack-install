package mojang

import (
	"github.com/nothub/mrpack-install/requester"
)

const manifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

func LatestVersion() (string, error) {
	var manifest struct {
		Latest struct {
			Release  string `json:"release"`
			Snapshot string `json:"snapshot"`
		} `json:"latest"`
	}
	err := requester.DefaultHttpClient.GetJson(manifestUrl, &manifest, nil)
	if err != nil {
		return "", err
	}
	return manifest.Latest.Release, nil
}
