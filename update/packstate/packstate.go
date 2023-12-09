package packstate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	modrinth "hub.lol/mrpack-install/modrinth/api"
	"hub.lol/mrpack-install/modrinth/mrpack"
)

const file = "packstate.json"

type Schema struct {
	Slug      string      `json:"slug"`
	ProjectId string      `json:"project-id"`
	Version   string      `json:"version"`
	VersionId string      `json:"version-id"`
	Deps      mrpack.Deps `json:"dependencies"`
	// Contains hashes of all downloads and override files of a state.
	Hashes map[string]modrinth.Hashes `json:"hashes"`
}

func (state *Schema) Save(serverDir string) error {
	if state == nil {
		return nil
	}

	j, err := json.Marshal(state)
	var buf bytes.Buffer
	err = json.Indent(&buf, j, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(serverDir, file), buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadPackState(serverDir string) (*Schema, error) {
	b, err := os.ReadFile(filepath.Join(serverDir, file))
	if err != nil {
		return nil, err
	}

	var state Schema
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

// FromArchive generates a Schema struct from a *.mrpack file.
func FromArchive(path string) (*Schema, error) {
	index, err := mrpack.ReadIndex(path)
	if err != nil {
		return nil, err
	}

	version, err := modrinth.Client.VersionFromMrpackFile(path)
	if err != nil {
		return nil, err
	}
	project, err := modrinth.Client.GetProject(version.ProjectId)
	if err != nil {
		return nil, err
	}

	var state Schema
	state.Slug = project.Slug
	state.ProjectId = project.Id
	state.Version = index.Version
	state.VersionId = version.Id
	state.Deps = index.Deps
	state.Hashes = make(map[string]modrinth.Hashes)

	// download hashes
	for _, indexFile := range index.Files {
		if indexFile.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}
		state.Hashes[indexFile.Path] = indexFile.Hashes
	}

	// override hashes
	for p, hashes := range mrpack.OverrideHashes(path) {
		state.Hashes[p] = hashes
	}

	return &state, nil
}
