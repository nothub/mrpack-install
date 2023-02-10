package update

import (
	"archive/zip"
	"crypto"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/util"
	"io"
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

func GetFileStrategy(hash string, path string) Strategy {
	if !util.PathIsFile(path) {
		return NoOp
	}
	match, _ := chksum.VerifyFile(path, hash, crypto.SHA1.New(), encoding.Hex)
	if match {
		return Delete
	} else {
		return Backup
	}
}

type Actions map[string]Strategy

// GetDeletionActions Three scenarios are possible:
// 1.File does not exist notice
// 2.File exists but hash value does not match, change the original file name to xxx.bak
// 3.File exists and the hash value matches
func GetDeletionActions(deletions *PackState, serverPath string) Actions {
	actions := make(Actions, 10)
	for filePath := range deletions.Hashes {
		t := GetFileStrategy(deletions.Hashes[filePath], filepath.Join(serverPath, string(filePath)))
		switch t {
		case Delete:
			fmt.Printf("[Delete]: %s \n", filePath)
			actions[filePath] = Delete
		case Backup:
			fmt.Printf("[Delete]: %s, The original file will be moved to updateBack folder\n", filePath)
			actions[filePath] = Backup
		}
	}
	return actions
}

// GetUpdateActions Three scenarios are possible:
// 1.File does not exist
// 2.File exists but hash value does not match, change the original file name to xxx.bak
// 3.File exists and the hash value matches, remove the item from the queue
func GetUpdateActions(updates *PackState, serverPath string) Actions {
	actions := make(Actions, 10)
	for filePath := range updates.Hashes {
		switch GetFileStrategy(updates.Hashes[filePath], filepath.Join(serverPath, string(filePath))) {
		case Delete:
			delete(updates.Hashes, filePath)
		case Backup:
			fmt.Printf("[Update]: %s ,The original file will be move to updateBack folder\n", filePath)
			actions[filePath] = Backup
		case NoOp:
			fmt.Printf("[Download]: %s \n", filePath)
			actions[filePath] = NoOp
		}
	}
	return actions
}

func ModPackDeleteDo(deleteList Actions, serverPath string) error {
	for filePath := range deleteList {
		switch deleteList[filePath] {
		case Delete:
			err := os.Remove(filepath.Join(serverPath, string(filePath)))
			if err != nil {
				return err
			}
		case Backup:
			err := os.MkdirAll(filepath.Dir(filepath.Join(serverPath, "updateBack", string(filePath))), 0755)
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(serverPath, string(filePath)), filepath.Join(serverPath, "updateBack", string(filePath)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ModPackUpdateDo(updateList Actions, updateFileInfo map[string]string, serverPath string, modPackPath string, downloadThreads int, retryTimes int) error {

	var downloads []*requester.Download
	downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, downloadThreads, retryTimes)

	// backup file and download file in modrinth index
	for filePath := range updateList {
		switch updateList[filePath] {
		case NoOp:
			if updateFileInfo[filePath].DownloadLink != nil {
				downloadPools.Downloads = append(downloadPools.Downloads, requester.NewDownload(updateFileInfo[filePath].DownloadLink, map[string]string{"sha1": updateFileInfo[filePath]}, filepath.Base(filePath), filepath.Join(serverPath, filepath.Dir(filePath))))
			}
		case Backup:
			err := os.MkdirAll(filepath.Dir(filepath.Join(serverPath, "updateBack", string(filePath))), 0755)
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(serverPath, string(filePath)), filepath.Join(serverPath, "updateBack", string(filePath)))
			if err != nil {
				return err
			}
			if updateFileInfo[filePath].DownloadLink != nil {
				downloadPools.Downloads = append(downloadPools.Downloads, requester.NewDownload(updateFileInfo[filePath].DownloadLink, map[string]string{"sha1": updateFileInfo[filePath]}, filepath.Base(filePath), filepath.Join(serverPath, filepath.Dir(filePath))))
			}
		}
	}
	downloadPools.Do()

	// unzip update file
	r, err := zip.OpenReader(modPackPath)
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

		if _, ok := updateFileInfo[filePathInZip]; ok && updateFileInfo[filePathInZip].DownloadLink == nil {

			targetPath := filepath.Join(serverPath, filePathInZip)

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
