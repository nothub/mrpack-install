package update

import (
	"archive/zip"
	"fmt"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/model"
	"github.com/nothub/mrpack-install/util"
	"strings"
)

func GenerateModPackInfo(modPackPatch string) (*model.ModPackInfo, error) {
	var modPackInfo model.ModPackInfo
	modPackInfo.File = make(map[string]model.FileInfo)

	modrinthIndex, err := mrpack.ReadIndex(modPackPatch)
	if err != nil {
		return nil, err
	}

	modPackInfo.Dependencies = modrinthIndex.Dependencies
	modPackInfo.ModPackVersion = modrinthIndex.VersionId
	modPackInfo.ModPackName = modrinthIndex.Name

	// Add modrinth.index file
	for _, file := range modrinthIndex.Files {
		modPackInfo.File[string(file.Hashes.Sha1)] = model.FileInfo{TargetPath: file.Path, DownloadLink: file.Downloads}
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
		modPackInfo.File[fileHash] = model.FileInfo{TargetPath: filePath}
	}

	return &modPackInfo, nil
}

// CompareModPackInfo Todo: Compare the two modPackInfo and generate a list of deletions and updates
func CompareModPackInfo(oldVersion *model.ModPackInfo, newVersion *model.ModPackInfo) (delete *model.ModPackInfo, add *model.ModPackInfo) {

	return nil, nil
}
