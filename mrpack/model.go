package mrpack

import "github.com/nothub/gorinth/api"

const indexFile string = "modrinth.index.json"

type Index struct {
	FormatVersion int          `json:"formatVersion"`
	Game          Game         `json:"game"`
	VersionId     string       `json:"versionId"`
	Name          string       `json:"name"`
	Summary       string       `json:"summary"`
	Files         []File       `json:"files"`
	Dependencies  Dependencies `json:"dependencies"`
}

type Game string

const (
	Minecraft Game = "minecraft"
)

type File struct {
	Path      string     `json:"path"`
	Hashes    api.Hashes `json:"hashes"`
	Env       Env        `json:"env"`
	Downloads []string   `json:"downloads"` // array of HTTPS URLs
	FileSize  int        `json:"fileSize"`  // size in bytes
}

type Env struct {
	Client api.Environment `json:"client"`
	Server api.Environment `json:"server"`
}

type Dependencies struct {
	Minecraft string `json:"minecraft"`
	Forge     string `json:"forge"`
	Fabric    string `json:"fabric-loader"`
	Quilt     string `json:"quilt-loader"`
}
