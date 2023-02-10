package mrpack

import (
	"archive/zip"
	"crypto"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/util"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func extractedPath(file *zip.File) (bool, string) {
	if strings.HasPrefix(file.Name, "overrides/") {
		return true, strings.TrimPrefix(file.Name, "overrides/")
	}
	if strings.HasPrefix(file.Name, "server-overrides/") {
		return true, strings.TrimPrefix(file.Name, "server-overrides/")
	}
	return false, ""
}

func GetOverrideHashes(zipFile string) (*map[string]string, error) {
	hashes := make(map[string]string)

	err := util.IterZip(zipFile, func(file *zip.File) error {
		ok, p := extractedPath(file)
		if !ok {
			// skip non-server overrides
			return nil
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		h, err := chksum.Create(reader, crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			return err
		}

		err = reader.Close()
		if err != nil {
			log.Println(err.Error())
		}

		hashes[p] = h

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
		_, err = io.Copy(outFile, fileReader)
		if err != nil {
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
