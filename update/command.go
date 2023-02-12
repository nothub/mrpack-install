package update

import (
	"fmt"
	"github.com/nothub/mrpack-install/cmd"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/util"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

func Cmd(opts *cmd.UpdateOpts, index *mrpack.Index, zipPath string) {
	fmt.Println("Updating:", index.Name)

	newState, err := BuildPackState(zipPath)
	if err != nil {
		log.Fatalln(err)
	}
	for filePath := range newState.Files {
		util.AssertPathSafe(filePath, opts.ServerDir)
	}

	oldState, err := LoadPackState(opts.ServerDir)
	if err != nil {
		log.Fatalln(err)
	}

	if !reflect.DeepEqual(oldState.Deps, newState.Deps) {
		// TODO: server update
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

	// handle old files
	for path := range oldState.Files {
		switch GetStrategy(oldState.Files[path], filepath.Join(opts.ServerDir, path)) {
		case Delete:
			err := os.Remove(filepath.Join(opts.ServerDir, path))
			if err != nil {
				log.Fatalln(err.Error())
			}
		case Backup:
			err := backup.Create(path, opts.ServerDir)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	}

	// new files
	var newFiles []File
	for path := range newState.Files {
		var f File
		f.Path = path
		switch GetStrategy(newState.Files[path], filepath.Join(opts.ServerDir, path)) {
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
