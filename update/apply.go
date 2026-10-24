package update

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/nothub/mrpack-install/web/download"
)

// stagingDir is where Stage downloads pack files, inside the server dir
// so moving them into place is a rename.
const stagingDir = ".mrpack-install-staging"

// Staging holds the pack files Stage downloaded, sorted by path.
type Staging struct {
	dir   string
	paths []string
}

// Stage downloads and hash-checks every index file the plan writes into
// the staging dir, before anything in the server dir changes. It first
// removes a staging dir a killed run left behind. On failure the staging
// dir is removed and the server dir is as it was.
func Stage(serverDir string, plan *Plan, index *mrpack.Index, archive string, dl download.Downloader) (*Staging, error) {
	dir := filepath.Join(serverDir, stagingDir)
	err := os.RemoveAll(dir)
	if err != nil {
		return nil, fmt.Errorf("removing old staging dir: %w", err)
	}

	writes := plan.writes()
	overrides := mrpack.OverrideHashes(archive)
	dl.Downloads = index.ServerDownloads(func(f mrpack.File) bool {
		_, overridden := overrides[f.Path]
		return writes[f.Path] && !overridden
	})
	staging := &Staging{dir: dir}
	for _, d := range dl.Downloads {
		staging.paths = append(staging.paths, d.Path)
	}
	slices.Sort(staging.paths)
	log.Printf("Downloading %v dependencies...\n", len(dl.Downloads))
	err = dl.Download(dir)
	if err != nil {
		if rerr := os.RemoveAll(dir); rerr != nil {
			log.Printf("Unable to remove staging dir: %s\n", rerr)
		}
		return nil, err
	}
	return staging, nil
}

// Apply runs plan on serverDir for the pack archive, with the files staging
// holds: it moves the files the plan backs up into backups, moves the
// staged files into place, extracts the overrides the plan writes, deletes
// the files it deletes, saves state, and removes the staging dir and the
// dirs those moves and deletes left empty.
func Apply(serverDir string, plan *Plan, staging *Staging, state *packstate.Schema, archive string, backups *backup.Dir) error {
	writes := plan.writes()
	var deletes []string
	var backedUp []string
	for _, a := range plan.Actions {
		switch a.Op {
		case OpBackupWrite, OpBackupDelete:
			backedUp = append(backedUp, a.Path)
		case OpDelete:
			deletes = append(deletes, a.Path)
		}
	}

	for _, p := range backedUp {
		err := backups.Move(serverDir, p)
		if err != nil {
			return fmt.Errorf("backing up %s: %w", p, err)
		}
	}

	for _, p := range staging.paths {
		dst, err := files.InDir(serverDir, p)
		if err != nil {
			return err
		}
		// A write over a file that still matches its old record replaces
		// it without a backup.
		err = os.Remove(dst)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("replacing %s: %w", p, err)
		}
		err = files.Move(filepath.Join(staging.dir, filepath.FromSlash(p)), dst)
		if err != nil {
			return fmt.Errorf("moving %s into place: %w", p, err)
		}
	}

	log.Println("Extracting overrides...")
	err := mrpack.ExtractOverrides(archive, serverDir, func(p string) bool { return writes[p] })
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

	// The sync is complete here, and the next run removes a leftover
	// staging dir anyway.
	err = os.RemoveAll(staging.dir)
	if err != nil {
		log.Printf("Unable to remove staging dir: %s\n", err)
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
