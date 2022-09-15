package model

import (
	"encoding/json"
	"github.com/nothub/mrpack-install/requester"
	"os"
	"path"
)

type ModPackInfo struct {
	File        *[]fileInfo `json:"file"`
	GameVersion string      `json:"gameVersion"`
	Loader      string      `json:"loader"`
	ModPackName string      `json:"modPackName"`
}

type fileInfo struct {
	FileHash     string   `json:"Hash"`
	FilePath     string   `json:"Path"`
	DownloadLink []string `json:"DownloadLink"`
	TargetPath   string   `json:"TargetPath"`
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

func WriteModPackInfo(modPack *ModPackInfo, modPackJsonFile string) error {
	if modPack != nil {
		modPackJsonByte, err := json.Marshal(modPack)
		if err != nil {
			return err
		}
		err = os.WriteFile(modPackJsonFile, modPackJsonByte, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (modPackInfo *ModPackInfo) GetDownloadPool(downloadPools *requester.DownloadPools) *requester.DownloadPools {
	for _, file := range *modPackInfo.File {
		if file.FilePath == "" && file.DownloadLink != nil {
			downloadPools.Downloads = append(downloadPools.Downloads, requester.NewDownload(file.DownloadLink, map[string]string{"sha1": file.FileHash}, path.Base(file.FilePath), path.Dir(file.FilePath)))
		}
	}
	return downloadPools
}
