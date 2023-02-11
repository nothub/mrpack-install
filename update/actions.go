package update

import (
	"archive/zip"
	"crypto"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/util"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Strategy uint8

const (
	Delete Strategy = iota
	Backup
	NoOp
)

// GetFileStrategy selects one of 3 strategies:
//
// 1. NoOp   - File does not exist
//
// 2. Delete - File exists and hash values match
//
// 3. Backup - File exists but hash values do not match
//
// Hash must be sha512 and hex encoded.
func GetFileStrategy(hash string, path string) Strategy {
	if !util.PathIsFile(path) {
		return NoOp
	}
	match, _ := chksum.VerifyFile(path, hash, crypto.SHA512.New(), encoding.Hex)
	if match {
		return Delete
	} else {
		return Backup
	}
}

type Actions map[string]Strategy

func GetDeletionActions(deletions *PackState, serverDir string) Actions {
	actions := make(Actions)
	for filePath := range deletions.Hashes {
		switch GetFileStrategy(deletions.Hashes[filePath], filepath.Join(serverDir, filePath)) {
		case Delete:
			actions[filePath] = Delete
		case Backup:
			actions[filePath] = Backup
		}
	}
	return actions
}

func GetUpdateActions(updates *PackState, serverDir string) Actions {
	actions := make(Actions)
	for filePath := range updates.Hashes {
		switch GetFileStrategy(updates.Hashes[filePath], filepath.Join(serverDir, filePath)) {
		case Delete:
			delete(updates.Hashes, filePath)
		case Backup:
			err := backup.Create(filePath, serverDir)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	}
	return actions
}

func HandleOldFiles(deletions Actions, serverDir string) error {
	for filePath, strategy := range deletions {
		switch strategy {
		case Delete:
			log.Println("Delete", filePath)
			err := os.Remove(filepath.Join(serverDir, filePath))
			if err != nil {
				return err
			}
		case Backup:
			err := backup.Create(filePath, serverDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Do(updates Actions, hashes map[string]string, serverDir string, zipPath string, threads int, retries int) error {
	var downloads []*requester.Download
	downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, threads, retries)

	// TODO: combine "updates Actions" and "hashes map[string]string" to a struct that also includes download links
	for filePath := range updates {
		downloadPools.Downloads = append(downloadPools.Downloads, requester.NewDownload(hashes[filePath].DownloadLink, map[string]string{"sha1": hashes[filePath]}, filepath.Base(filePath), filepath.Join(serverDir, filepath.Dir(filePath))))
	}
	downloadPools.Do()

	// unzip update file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(r)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		filePathInZip := f.Name
		if strings.HasPrefix(filePathInZip, "overrides/") {
			filePathInZip = strings.TrimPrefix(filePathInZip, "overrides/")
		} else if strings.HasPrefix(filePathInZip, "server-overrides/") {
			filePathInZip = strings.TrimPrefix(filePathInZip, "server-overrides/")
		} else {
			continue
		}

		if _, ok := hashes[filePathInZip]; ok && hashes[filePathInZip].DownloadLink == nil {

			targetPath := filepath.Join(serverDir, filePathInZip)

			err := os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return err
			}

			fileReader, err := f.Open()
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

		}
	}
	return nil
}