package update

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"os"
	"reflect"
	"strings"
)

type PackState struct {
	Dependencies mrpack.Dependencies `json:"dependencies"`
	Hashes       map[string]string   `json:"hashes"` // Hex encoded SHA512 checksums
}

func (state *PackState) Save(path string) error {
	if state == nil {
		return nil
	}

	j, err := json.Marshal(state)
	var buf bytes.Buffer
	err = json.Indent(&buf, j, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadPackState(path string) (*PackState, error) {
	b, err := os.ReadFile(path)
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

func BuildPackState(zipPath string) (*PackState, error) {
	index, err := mrpack.ReadIndex(zipPath)
	if err != nil {
		return nil, err
	}

	var state PackState
	state.Dependencies = index.Dependencies

	// mrpack index files (downloads)
	for _, file := range index.Files {
		if file.Env.Server == api.UnsupportedEnvSupport {
			continue
		}
		state.Hashes[file.Path] = file.Hashes.Sha512
	}

	// override files
	state.Hashes = mrpack.OverrideHashes(zipPath)

	return &state, nil
}

func CompareModPackInfo(old PackState, new PackState) (deletions *PackState, updates *PackState, err error) {
	if !reflect.DeepEqual(old.Dependencies, new.Dependencies) {
		// TODO: server update
		return nil, nil, errors.New("mismatched versions, please upgrade manually")
	}

	for path := range old.Hashes {
		// ignore unchanged files
		if new.Hashes[path] == old.Hashes[path] {
			delete(old.Hashes, path)
			delete(new.Hashes, path)
		}

		// skip deletion of old files that we overwrite with new files
		if _, found := new.Hashes[path]; found {
			delete(old.Hashes, path)
		}
	}

	return &old, &new, nil
}
