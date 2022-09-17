package model

import (
	"bytes"
	"encoding/json"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"os"
)

type Path string
type FileMap map[Path]FileInfo

type ModPackInfo struct {
	ModPackVersion string              `json:"modPackVersion"`
	ModPackName    string              `json:"modPackName"`
	File           FileMap             `json:"file"`
	Dependencies   mrpack.Dependencies `json:"dependencies"`
}

type FileInfo struct {
	Hash         string   `json:"Hash"`
	DownloadLink []string `json:"DownloadLink"`
}

func ReadModPackInfo(modPackJsonFile string) (*ModPackInfo, error) {

	var modPackJsonByte []byte
	var err error
	modPackJsonByte, err = os.ReadFile(modPackJsonFile)
	if err != nil {
		return nil, err
	}

	modPackJson := ModPackInfo{}
	err = json.Unmarshal(modPackJsonByte, &modPackJson)

	if err != nil {
		return nil, err
	}
	return &modPackJson, nil
}

func (modPackInfo *ModPackInfo) Write(modPackJsonFile string) error {
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
