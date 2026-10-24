package update

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/nothub/mrpack-install/web/download"
)

// Apply runs plan on serverDir for the pack with index and archive: it
// moves the files the plan backs up into backups, downloads and extracts
// the files it writes, deletes the files it deletes, saves state, and
// removes the dirs those moves and deletes left empty.
func Apply(serverDir string, plan *Plan, state *packstate.Schema, index *mrpack.Index, archive string, dl download.Downloader, backups *backup.Dir) error {
	writes := make(map[string]bool)
	var deletes []string
	var backedUp []string
	for _, a := range plan.Actions {
		switch a.Op {
		case OpWrite:
			writes[a.Path] = true
		case OpBackupWrite:
			writes[a.Path] = true
			backedUp = append(backedUp, a.Path)
		case OpDelete:
			deletes = append(deletes, a.Path)
		case OpBackupDelete:
			backedUp = append(backedUp, a.Path)
		}
	}

	for _, p := range backedUp {
		err := backups.Move(serverDir, p)
		if err != nil {
			return fmt.Errorf("backing up %s: %w", p, err)
		}
	}

	overrides := mrpack.OverrideHashes(archive)
	dl.Downloads = index.ServerDownloads(func(f mrpack.File) bool {
		_, overridden := overrides[f.Path]
		return writes[f.Path] && !overridden
	})
	log.Printf("Downloading %v dependencies...\n", len(dl.Downloads))
	err := dl.Download(serverDir)
	if err != nil {
		return err
	}

	log.Println("Extracting overrides...")
	err = mrpack.ExtractOverrides(archive, serverDir, func(p string) bool { return writes[p] })
	if err != nil {
		return fmt.Errorf("extracting overrides: %w", err)
	}

	for _, p := range deletes {
		full, err := files.InDir(serverDir, p)
		if err != nil {
			return err
		}
		err = os.Remove(full)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("deleting %s: %w", p, err)
		}
		log.Printf("Deleted: %s\n", p)
	}

	err = state.Save(serverDir)
	if err != nil {
		return fmt.Errorf("saving %s: %w", packstate.FileName, err)
	}

	removeEmptiedDirs(serverDir, append(deletes, backedUp...))
	return nil
}

// removeEmptiedDirs removes the parent dirs of the deleted paths that are
// empty now, innermost first, up to but excluding serverDir. A dir that
// still holds anything, or a symlink, which os.Remove would unlink whatever
// its target holds, stops the walk up.
func removeEmptiedDirs(serverDir string, deleted []string) {
	for _, p := range deleted {
		dir := filepath.Dir(filepath.Join(serverDir, filepath.FromSlash(p)))
		for len(dir) > len(serverDir) {
			fi, err := os.Lstat(dir)
			if err != nil || !fi.IsDir() {
				break
			}
			if os.Remove(dir) != nil {
				break
			}
			log.Printf("Deleted empty dir: %s\n", dir)
			dir = filepath.Dir(dir)
		}
	}
}
