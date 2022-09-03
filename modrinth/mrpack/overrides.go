package mrpack

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ExtractOverrides(zipFile string, target string) error {
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
		if file.FileInfo().IsDir() {
			continue
		}

		filePath := file.Name
		if strings.HasPrefix(filePath, "overrides/") {
			filePath = strings.TrimPrefix(filePath, "overrides/")
		} else if strings.HasPrefix(filePath, "server-overrides/") {
			filePath = strings.TrimPrefix(filePath, "server-overrides/")
		} else {
			continue
		}

		targetPath := path.Join(target, filePath)

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

		log.Println("Override file extracted:", targetPath)
	}

	return nil
}
