package update

import (
	"crypto"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"hub.lol/mrpack-install/files"
	"hub.lol/mrpack-install/modrinth/mrpack"
	"hub.lol/mrpack-install/update/backup"
	"hub.lol/mrpack-install/update/packstate"
	"hub.lol/mrpack-install/web/download"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

import "golang.org/x/exp/slices"

func Cmd(serverDir string, dlThreads uint8, dlRetries uint8, index *mrpack.Index, zipPath string, oldState *packstate.Schema) {
	fmt.Printf("Updating %q in %q with %q\n", index.Name, serverDir, zipPath)
	err := os.Chdir(serverDir)
	if err != nil {
		log.Fatalln(err)
	}

	newState, err := packstate.FromArchive(zipPath)
	if err != nil {
		log.Fatalln(err)
	}
	for filePath := range newState.Hashes {
		files.AssertSafe(filepath.Join(serverDir, filePath), serverDir)
	}

	if !reflect.DeepEqual(oldState.Deps, newState.Deps) {
		// TODO: better message
		log.Fatalln("mismatched versions, please upgrade manually")
	}

	// ignore files that are left unchanged in the update process
	var ignores []string
	for path := range newState.Hashes {
		// check if file exists
		if !files.IsFile(path) {
			continue
		}
		// check if pack file changes in update
		if newState.Hashes[path] != oldState.Hashes[path] {
			continue
		}
		// check if local file changes in update
		currentHash, err := chksum.FromFile(path, crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			log.Fatalln(err)
		}
		if currentHash == newState.Hashes[path].Sha512 {
			ignores = append(ignores, path)
		}
	}

	// backup if the file exists but the new hash value does not match
	for path := range oldState.Hashes {
		if slices.Contains(ignores, path) {
			continue
		}

		if !files.IsFile(path) {
			continue
		}

		// check if file will be replaced
		_, ok := newState.Hashes[path]
		if !ok {
			continue
		}

		// TODO: too many backups check hashes, combine with ignores list?
		err := backup.Create(path, serverDir)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	// downloads
	var downloads []*download.Download
	for _, dl := range index.ServerDownloads() {
		if !slices.Contains(ignores, dl.Path) {
			downloads = append(downloads, dl)
		}
	}

	fmt.Printf("Downloading %v dependencies...\n", len(downloads))
	downloader := download.Downloader{
		Downloads: downloads,
		Threads:   int(dlThreads),
		Retries:   int(dlRetries),
	}
	downloader.Download(serverDir)

	// overrides
	fmt.Println("Extracting overrides...")
	err = mrpack.ExtractOverrides(zipPath, serverDir)
	if err != nil {
		log.Fatalln(err)
	}

	// save state file
	err = newState.Save(serverDir)
	if err != nil {
		log.Fatalln(err)
	}

	files.RmEmptyDirs(serverDir)

	fmt.Println("Update finished :) Have a nice day ✌️")
}
