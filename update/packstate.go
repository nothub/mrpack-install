package update

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"strings"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
)

const statefile = "packstate.json"

type PackState struct {
	Name    string                     `json:"name"`
	Version string                     `json:"version"`
	Deps    mrpack.Deps                `json:"dependencies"`
	Files   map[string]modrinth.Hashes `json:"files"`
}

func (state *PackState) Save(serverDir string) error {
	if state == nil {
		return nil
	}

	j, err := json.Marshal(state)
	var buf bytes.Buffer
	err = json.Indent(&buf, j, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(serverDir, statefile), buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadPackState(serverDir string) (*PackState, error) {
	b, err := os.ReadFile(path.Join(serverDir, statefile))
	if err != nil {
		return nil, err
	}

	var state PackState
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func BuildPackState(index *mrpack.Index, zipPath string) (*PackState, error) {
	var state PackState
	state.Name = index.Name
	state.Version = index.Version
	state.Deps = index.Deps
	state.Files = make(map[string]modrinth.Hashes)

	// downloads
	for _, indexFile := range index.Files {
		if indexFile.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}
		state.Files[indexFile.Path] = indexFile.Hashes
	}

	// overrides
	for p, hashes := range mrpack.OverrideHashes(zipPath) {
		state.Files[p] = hashes
	}

	return &state, nil
}
