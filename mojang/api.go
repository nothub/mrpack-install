package mojang

import (
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"github.com/nothub/mrpack-install/web"
	"time"
)

const manifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

var playerUrl = func(name string) string {
	return "https://api.mojang.com/users/profiles/minecraft/" + name
}

type Manifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		Id          string    `json:"id"`
		Type        string    `json:"type"`
		Url         string    `json:"url"`
		Time        time.Time `json:"time"`
		ReleaseTime time.Time `json:"releaseTime"`
	} `json:"versions"`
}

type Meta struct {
	Arguments struct {
		Game []interface{} `json:"game"`
		Jvm  []interface{} `json:"jvm"`
	} `json:"arguments"`
	AssetIndex struct {
		Id        string `json:"id"`
		Sha1      string `json:"sha1"`
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		Url       string `json:"url"`
	} `json:"assetIndex"`
	Assets          string `json:"assets"`
	ComplianceLevel int    `json:"complianceLevel"`
	Downloads       struct {
		Client struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			Url  string `json:"url"`
		} `json:"client"`
		ClientMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			Url  string `json:"url"`
		} `json:"client_mappings"`
		Server struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			Url  string `json:"url"`
		} `json:"server"`
		ServerMappings struct {
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			Url  string `json:"url"`
		} `json:"server_mappings"`
	} `json:"downloads"`
	Id          string `json:"id"`
	JavaVersion struct {
		Component    string `json:"component"`
		MajorVersion int    `json:"majorVersion"`
	} `json:"javaVersion"`
	Libraries []struct {
		Downloads struct {
			Artifact struct {
				Path string `json:"path"`
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				Url  string `json:"url"`
			} `json:"artifact"`
		} `json:"downloads"`
		Name  string `json:"name"`
		Rules []struct {
			Action string `json:"action"`
			Os     struct {
				Name string `json:"name"`
			} `json:"os"`
		} `json:"rules,omitempty"`
	} `json:"libraries"`
	Logging struct {
		Client struct {
			Argument string `json:"argument"`
			File     struct {
				Id   string `json:"id"`
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
				Url  string `json:"url"`
			} `json:"file"`
			Type string `json:"type"`
		} `json:"client"`
	} `json:"logging"`
	MainClass              string    `json:"mainClass"`
	MinimumLauncherVersion int       `json:"minimumLauncherVersion"`
	ReleaseTime            time.Time `json:"releaseTime"`
	Time                   time.Time `json:"time"`
	Type                   string    `json:"type"`
}

type Player struct {
	Uuid string `json:"id"`
	Name string `json:"name"`
}

func GetManifest() (*Manifest, error) {
	var manifest Manifest
	err := web.DefaultClient.GetJson(manifestUrl, &manifest, nil)
	if err != nil {
		return nil, err
	}
	return &manifest, nil
}

func LatestRelease() (string, error) {
	manifest, err := GetManifest()
	if err != nil {
		return "", err
	}
	return manifest.Latest.Release, nil
}

func GetMeta(version string) (*Meta, error) {
	var manifest Manifest
	err := web.DefaultClient.GetJson(manifestUrl, &manifest, nil)
	if err != nil {
		return nil, err
	}

	u := ""
	for i := range manifest.Versions {
		v := manifest.Versions[i]
		if v.Id == version {
			u = v.Url
			break
		}
	}

	if u == "" {
		return nil, errors.New("no manifest entry found for version")
	}

	var meta Meta
	err = web.DefaultClient.GetJson(u, &meta, nil)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func GetPlayer(name string) (*Player, error) {
	// TODO: why is this here? i can not imagine
	//  what i thought this could be useful for 🤔
	var player Player
	err := web.DefaultClient.GetJson(playerUrl(name), &player, nil)
	if err != nil {
		return nil, err
	}

	id, err := formatUuid(player.Uuid)
	if err != nil {
		return nil, err
	}
	player.Uuid = id

	return &player, nil
}

func formatUuid(s string) (string, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}

	id, err := uuid.FromBytes(b)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
