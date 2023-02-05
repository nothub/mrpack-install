package update

import (
	"archive/zip"
	"errors"
	"fmt"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/util"
	"reflect"
	"strings"
)

func GenerateModPackInfo(modrinthIndex *mrpack.Index) (*ModPackInfo, error) {
	var modPackInfo ModPackInfo
	modPackInfo.Hashes = make(util.Hashes)

	modPackInfo.Dependencies = modrinthIndex.Dependencies
	modPackInfo.PackVersion = modrinthIndex.VersionId
	modPackInfo.PackName = modrinthIndex.Name

	// Add modrinth.index file
	for _, file := range modrinthIndex.Files {
		if file.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}
		modPackInfo.Hashes[file.Path] = string(file.Hashes.Sha1)
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
		modPackInfo.Hashes[filePath] = fileHash
	}

	return &modPackInfo, nil
}

func CompareModPackInfo(old ModPackInfo, new ModPackInfo) (deleteFileInfo *ModPackInfo, updateFileInfo *ModPackInfo, err error) {
	if old.PackName != new.PackName || !reflect.DeepEqual(old.Dependencies, new.Dependencies) {
		return nil, nil, errors.New("for mismatched versions, please upgrade manually")
	}

	for path := range old.Hashes {
		// ignore unchanged files
		if new.Hashes[path] == old.Hashes[path] {
			delete(old.Hashes, path)
			delete(new.Hashes, path)
		}

		// do not delete old files that we overwrite with new files
		if _, found := new.Hashes[path]; found {
			delete(old.Hashes, path)
		}
	}

	return &old, &new, nil
}
