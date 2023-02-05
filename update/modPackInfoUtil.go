package update

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/util"
	"reflect"
	"strings"
)

func GenerateModPackInfo(modPackPatch string) (*ModPackInfo, error) {
	var modPackInfo ModPackInfo
	modPackInfo.File = make(FileMap)

	modrinthIndex, err := mrpack.ReadIndex(modPackPatch)
	if err != nil {
		return nil, err
	}

	modPackInfo.Dependencies = modrinthIndex.Dependencies
	modPackInfo.ModPackVersion = modrinthIndex.VersionId
	modPackInfo.ModPackName = modrinthIndex.Name

	// Add modrinth.index file
	for _, file := range modrinthIndex.Files {
		if file.Env.Server == "unsupported" {
			continue
		}
		modPackInfo.File[Path(file.Path)] = FileInfo{Hash: string(file.Hashes.Sha1), DownloadLink: file.Downloads}
	}

	// Add overrides file
	r, err := zip.OpenReader(modPackPatch)
	if err != nil {
		return nil, err
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

		filePath := f.Name
		if strings.HasPrefix(filePath, "overrides/") {
			filePath = strings.TrimPrefix(filePath, "overrides/")
		} else if strings.HasPrefix(filePath, "server-overrides/") {
			filePath = strings.TrimPrefix(filePath, "server-overrides/")
		} else {
			continue
		}

		readCloser, err := f.Open()
		if err != nil {
			return nil, err
		}

		fileHash, err := util.GetReadCloserSha1(readCloser)
		if err != nil {
			return nil, err
		}
		err = readCloser.Close()
		if err != nil {
			return nil, err
		}
		modPackInfo.File[Path(filePath)] = FileInfo{Hash: fileHash}
	}

	return &modPackInfo, nil
}

func CompareModPackInfo(old ModPackInfo, new ModPackInfo) (deleteFileInfo *ModPackInfo, updateFileInfo *ModPackInfo, err error) {
	if old.ModPackName != new.ModPackName || !reflect.DeepEqual(old.Dependencies, new.Dependencies) {
		return nil, nil, errors.New("for mismatched versions, please upgrade manually")
	}

	for path := range old.File {
		// ignore unchanged files
		if new.File[path].Hash == old.File[path].Hash {
			delete(old.File, path)
			delete(new.File, path)
		}

		// do not delete old files that we overwrite with new files
		if _, found := new.File[path]; found {
			delete(old.File, path)
		}
	}

	return &old, &new, nil
}
