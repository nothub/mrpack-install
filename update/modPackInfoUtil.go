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

	modrinthIndex, err := mrpack.ReadIndex(modPackPatch)
	if err != nil {
		return nil, err
	}

	modPackInfo.Dependencies = modrinthIndex.Dependencies
	modPackInfo.ModPackVersion = modrinthIndex.VersionId
	modPackInfo.ModPackName = modrinthIndex.Name

	// Add modrinth.index file
	for _, file := range modrinthIndex.Files {
		var tmpFileInfo model.FileInfo
		tmpFileInfo.FileHash = string(file.Hashes.Sha1)
		tmpFileInfo.TargetPath = file.Path
		tmpFileInfo.DownloadLink = file.Downloads

		modPackInfo.File = append(modPackInfo.File, tmpFileInfo)
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

		var tmpFileInfo model.FileInfo

		readCloser, err := f.Open()
		if err != nil {
			return nil, err
		}

		tmpFileInfo.FileHash, err = util.GetReadCloserSha1(readCloser)
		if err != nil {
			return nil, err
		}
		tmpFileInfo.TargetPath = filePath

		modPackInfo.File = append(modPackInfo.File, tmpFileInfo)
	}

	return &modPackInfo, nil
}

// CompareModPackInfo Todo: Compare the two modPackInfo and generate a list of deletions and updates
func CompareModPackInfo(oldVersion *model.ModPackInfo, newVersion *model.ModPackInfo) (delete *model.ModPackInfo, add *model.ModPackInfo) {

	return nil, nil
}
