package update

import (
	"bytes"
	"encoding/json"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"os"
)

type PackState struct {
	Schema       uint8               `json:"schema_version"`
	PackVersion  string              `json:"pack_version"`
	PackName     string              `json:"pack_name"`
	Hashes       map[string]string   `json:"hashes"`
	Dependencies mrpack.Dependencies `json:"dependencies"`
}

type FileInfo struct {
	Hash          string   `json:"hash"`
	DownloadLinks []string `json:"download_links"`
}

func ReadModPackInfo(modPackJsonFile string) (*PackState, error) {

	var modPackJsonByte []byte
	var err error
	modPackJsonByte, err = os.ReadFile(modPackJsonFile)
	if err != nil {
		return nil, err
	}

	modPackJson := PackState{}
	err = json.Unmarshal(modPackJsonByte, &modPackJson)

	if err != nil {
		return nil, err
	}
	return &modPackJson, nil
}

func (modPackInfo *PackState) Write(modPackJsonFile string) error {
	if modPackInfo != nil {
		modPackJsonByte, err := json.Marshal(modPackInfo)
		var out bytes.Buffer
		err = json.Indent(&out, modPackJsonByte, "", "\t")
		if err != nil {
			return err
		}
		err = os.WriteFile(modPackJsonFile, out.Bytes(), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
