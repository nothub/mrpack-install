package model

import (
	"bytes"
	"encoding/json"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/requester"
	"os"
	"path"
)

type ModPackInfo struct {
	ModPackVersion string              `json:"modPackVersion"`
	ModPackName    string              `json:"modPackName"`
	File           []FileInfo          `json:"file"`
	Dependencies   mrpack.Dependencies `json:"dependencies"`
}

type FileInfo struct {
	FileHash     string   `json:"Hash"`
	TargetPath   string   `json:"TargetPath"`
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

func (modPackInfo *ModPackInfo) GetDownloadPool(downloadPools *requester.DownloadPools) *requester.DownloadPools {
	for _, file := range modPackInfo.File {
		if file.DownloadLink != nil {
			downloadPools.Downloads = append(downloadPools.Downloads, requester.NewDownload(file.DownloadLink, map[string]string{"sha1": file.FileHash}, path.Base(file.TargetPath), path.Dir(file.TargetPath)))
		}
	}
	return downloadPools
}
