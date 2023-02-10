package update

import (
	"archive/zip"
	"crypto"
	"errors"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"reflect"
	"strings"
)

func GenerateModPackInfo(zipFile string) (*PackState, error) {
	index, err := mrpack.ReadIndex(zipFile)
	if err != nil {
		return nil, err
	}

	var state PackState
	state.PackName = index.Name
	state.PackVersion = index.VersionId
	state.Dependencies = index.Dependencies

	hashes, err := mrpack.GetOverrideHashes(zipFile)
	if err != nil {
		return nil, err
	}
	state.Hashes = *hashes

	// Add modrinth.index file
	for _, file := range index.Files {
		if file.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}
		state.Hashes[file.Path] = string(file.Hashes.Sha1)
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

		fileHash, err := chksum.Create(readCloser, crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			return nil, err
		}
		err = readCloser.Close()
		if err != nil {
			return nil, err
		}
		state.Hashes[filePath] = fileHash
	}

	return &state, nil
}

func CompareModPackInfo(old PackState, new PackState) (deleteFileInfo *PackState, updateFileInfo *PackState, err error) {
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
