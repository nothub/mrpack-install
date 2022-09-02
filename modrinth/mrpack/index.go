package mrpack

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

import modrinth "github.com/nothub/gorinth/modrinth/api"

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
	Path      string          `json:"path"`
	Hashes    modrinth.Hashes `json:"hashes"`
	Env       Env             `json:"env"`
	Downloads []string        `json:"downloads"` // array of HTTPS URLs
	FileSize  int             `json:"fileSize"`  // size in bytes
}

type Env struct {
	Client modrinth.EnvSupport `json:"client"`
	Server modrinth.EnvSupport `json:"server"`
}

type Dependencies struct {
	Minecraft string `json:"minecraft"`
	Forge     string `json:"forge"`
	Fabric    string `json:"fabric-loader"`
	Quilt     string `json:"quilt-loader"`
}

func ReadIndex(zipFile string) (*Index, error) {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer func(zipReader *zip.ReadCloser) {
		err := zipReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(zipReader)

	var indexFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == "modrinth.index.json" {
			indexFile = file
			break
		}
	}
	if indexFile == nil {
		return nil, errors.New("no index file found")
	}

	fileReader, err := indexFile.Open()
	if err != nil {
		return nil, err
	}
	defer func(fileReader io.ReadCloser) {
		err := fileReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(fileReader)

	var index Index
	err = json.NewDecoder(fileReader).Decode(&index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}
