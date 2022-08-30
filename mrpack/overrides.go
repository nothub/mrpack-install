package mrpack

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type Side string

const (
	Client Side = "client"
	Server Side = "server"
)

func ExtractOverrides(zipFile string, target string, side Side) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func(zipReader *zip.ReadCloser) {
		err := zipReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(zipReader)

	for _, file := range zipReader.File {
		if file.Name == "modrinth.index.json" {
			continue
		}

		switch side {
		case Client:
			//"overrides", "client-overrides"
		case Server:
			//"overrides", "server-overrides"
		default:
			//illegal
		}

		// create parent directory tree
		targetPath := path.Join(target, file.Name)
		err := os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			return err
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		var buf []byte
		_, err = fileReader.Read(buf)
		if err != nil {
			return err
		}
		err = os.WriteFile(targetPath, buf, 0644)
		if err != nil {
			return err
		}

		err = fileReader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
