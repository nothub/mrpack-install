package update

import (
	"fmt"
	"github.com/nothub/mrpack-install/cmd"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/util"
	"log"
	"path/filepath"
	"reflect"
)

func Cmd(opts *cmd.UpdateOpts, index *mrpack.Index, zipPath string) {
	fmt.Printf("Updating %s with %s", opts.ServerDir, zipPath)

	oldState, err := LoadPackState(opts.ServerDir)
	if err != nil {
		log.Fatalln(err)
	}

	newState, err := BuildPackState(index, zipPath)
	if err != nil {
		log.Fatalln(err)
	}
	for filePath := range newState.Files {
		util.AssertPathSafe(filePath, opts.ServerDir)
	}

	if !reflect.DeepEqual(oldState.Deps, newState.Deps) {
		// TODO: better message
		log.Fatalln("mismatched versions, please upgrade manually")
	}

	for path := range oldState.Files {
		// ignore unchanged files
		if newState.Files[path] == oldState.Files[path] {
			delete(oldState.Files, path)
			delete(newState.Files, path)
		}
		// skip deletion of old files that we overwrite with new files
		if _, found := newState.Files[path]; found {
			delete(oldState.Files, path)
		}
	}

	for path, hashes := range oldState.Files {
		if ShouldBackup(path, hashes.Sha512) {
			err := backup.Create(path, opts.ServerDir)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	}

	// TODO: correctly handle new files
	var newFiles []mrpack.File
	for path, hashes := range newState.Files {
		var f mrpack.File
		f.Path = path
		switch GetStrategy(hashes.Sha512, filepath.Join(opts.ServerDir, path)) {
		case Delete:
			delete(newState.Files, path)
		case Backup:
			err := backup.Create(path, opts.ServerDir)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
		newFiles = append(newFiles, f)
	}

	err = Do(newFiles, opts.ServerDir, zipPath, opts.DownloadThreads, opts.RetryTimes)
	if err != nil {
		log.Fatalln(err)
	}

	util.RemoveEmptyDirs(opts.ServerDir)

	err = newState.Save(opts.ServerDir)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Done :) Have a nice day ✌️")
}
