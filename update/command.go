package update

import (
	"crypto"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/files"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/nothub/mrpack-install/web/download"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
)

func Cmd(serverDir string, dlThreads uint8, dlRetries uint8, index *mrpack.Index, zipPath string, oldState *packstate.Schema, optionalSelected []string, optionalDisableAll bool) {
	log.Printf("Updating %q in %q with %q\n", index.Name, serverDir, zipPath)
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
	downloads := index.ServerDownloads(func(f mrpack.File) bool {
		if slices.Contains(ignores, f.Path) {
			return false
		}
		switch f.Env.Server {
		case modrinth.RequiredEnvSupport:
			return true
		case modrinth.OptionalEnvSupport:
			if optionalDisableAll {
				return false
			}
			if len(optionalSelected) < 1 {
				return true
			}
			for _, p := range optionalSelected {
				if regexp.MustCompile(p).Match([]byte(filepath.Base(f.Path))) {
					return true
				}
			}
			return false
		case modrinth.UnsupportedEnvSupport:
			return false
		default:
			log.Fatalf("Unexpected environment configuration %q for mod %q\n", f.Env.Server, f.Path)
			return false
		}
	})
	log.Printf("Downloading %v dependencies...\n", len(downloads))
	downloader := download.Downloader{
		Downloads: downloads,
		Threads:   int(dlThreads),
		Retries:   int(dlRetries),
	}
	downloader.Download(serverDir)

	// overrides
	log.Println("Extracting overrides...")
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

	log.Println("Update finished :) Have a nice day ✌️")
}
