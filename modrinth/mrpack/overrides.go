package mrpack

import (
	"archive/zip"
	"fmt"
	"github.com/nothub/mrpack-install/util"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GetOverrideHashes(zipFile string) (*util.Hashes, error) {
	hashes := make(util.Hashes)

	err := util.IterZip(zipFile, func(file *zip.File) error {
		var name string
		if strings.HasPrefix(file.Name, "overrides/") {
			name = strings.TrimPrefix(file.Name, "overrides/")
		} else if strings.HasPrefix(file.Name, "server-overrides/") {
			name = strings.TrimPrefix(file.Name, "server-overrides/")
		} else {
			return nil
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		hash, err := util.GetReadCloserSha1(reader)
		if err != nil {
			return err
		}

		err = reader.Close()
		if err != nil {
			return err
		}

		hashes[name] = hash

		return nil
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	return &hashes, nil
}

func ExtractOverrides(zipFile string, target string) error {
	err := util.IterZip(zipFile, func(file *zip.File) error {

		filePath := file.Name
		if strings.HasPrefix(filePath, "overrides/") {
			filePath = strings.TrimPrefix(filePath, "overrides/")
		} else if strings.HasPrefix(filePath, "server-overrides/") {
			filePath = strings.TrimPrefix(filePath, "server-overrides/")
		} else {
			return nil
		}

		targetPath := path.Join(target, filePath)
		ok, err := util.PathIsSubpath(targetPath, target)
		if err != nil {
			log.Println(err.Error())
		}
		if err != nil || !ok {
			log.Fatalln("File path is not safe: " + targetPath)
		}

		err = os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			return err
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, fileReader); err != nil {
			return err
		}

		err = fileReader.Close()
		if err != nil {
			return err
		}
		err = outFile.Close()
		if err != nil {
			return err
		}

		fmt.Println("Override file extracted:", targetPath)

		return nil
	})

	if err != nil {
		log.Fatalln(err.Error())
	}

	return nil
}
