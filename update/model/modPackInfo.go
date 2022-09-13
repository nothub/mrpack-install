package model

import (
	"encoding/json"
	"os"
)

type modPackInfo struct {
	File        *[]fileInfo `json:"file"`
	GameVersion string      `json:"gameVersion"`
	Loader      string      `json:"loader"`
	ModPackName string      `json:"modPackName"`
}

type fileInfo struct {
	FileHash string `json:"Hash"`
	FilePath string `json:"Path"`
}

func ReadModPackInfo(modPackJsonFile string) (*modPackInfo, error) {

	var modPackJsonByte []byte
	var err error
	modPackJsonByte, err = os.ReadFile(modPackJsonFile)
	if err != nil {
		return nil, err
	}

	modPackJson := modPackInfo{}
	err = json.Unmarshal(modPackJsonByte, &modPackJson)

	if err != nil {
		return nil, err
	}
	return &modPackJson, nil
}

func WriteModPackInfo(modPack *modPackInfo, modPackJsonFile string) error {
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
