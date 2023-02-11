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

func extPath(file *zip.File) (bool, string) {
	if strings.HasPrefix(file.Name, "overrides/") {
		return true, strings.TrimPrefix(file.Name, "overrides/")
	}
	if strings.HasPrefix(file.Name, "server-overrides/") {
		return true, strings.TrimPrefix(file.Name, "server-overrides/")
	}
	return false, ""
}

func ExtractOverrides(zipFile string, serverDir string) error {
	err := util.IterZip(zipFile, func(file *zip.File) error {
		ok, filePath := extPath(file)
		if !ok {
			// skip non-server override files
			return nil
		}

		util.AssertPathSafe(filePath, serverDir)
		targetPath := path.Join(serverDir, filePath)

		err := os.MkdirAll(filepath.Dir(targetPath), 0755)
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

func OverrideHashes(zipFile string) map[string]string {
	hashes := make(map[string]string)

	err := util.IterZip(zipFile, func(file *zip.File) error {
		ok, p := extPath(file)
		if !ok {
			// skip non-server override files
			return nil
		}

		r, err := file.Open()
		if err != nil {
			return err
		}

		h, err := chksum.Create(r, crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			return err
		}
		hashes[p] = h

		err = r.Close()
		if err != nil {
			log.Println(err.Error())
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	return hashes
}
