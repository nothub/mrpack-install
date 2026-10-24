package update

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/nothub/mrpack-install/files"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
)

// Installer installs a server and returns the jar that starts it, as
// server.Installer does.
type Installer interface {
	Install(serverDir string, serverFile string) (string, error)
}

// InstallServer runs inst when the plan installs the server and records
// the server file it produced in state. Without an install, the old record
// carries over. A recorded server file that no longer matches its record,
// or an unrecorded file at the --server-file name, is backed up before the
// install. A recorded file the install replaced under another name is
// deleted, so a start command naming an old versioned jar fails instead of
// starting the old loader.
func InstallServer(serverDir string, plan *Plan, old *packstate.Schema, state *packstate.Schema, inst Installer, serverFile string, backups *backup.Dir) error {
	var recorded *packstate.ServerFile
	if old != nil {
		recorded = old.ServerFile
	}
	if !plan.InstallServer {
		state.ServerFile = recorded
		return nil
	}

	var recordedPath string
	if recorded != nil && recorded.Path != "" {
		var err error
		recordedPath, err = files.InDir(serverDir, recorded.Path)
		if err != nil {
			return err
		}
		err = backUpUnless(serverDir, recorded.Path, recordedPath, &recorded.Hashes, backups)
		if err != nil {
			return err
		}
	}
	if serverFile != "" && (recorded == nil || recorded.Path != serverFile) {
		target, err := files.InDir(serverDir, serverFile)
		if err != nil {
			return err
		}
		err = backUpUnless(serverDir, serverFile, target, nil, backups)
		if err != nil {
			return err
		}
	}

	file, err := inst.Install(serverDir, serverFile)
	if err != nil {
		return err
	}

	// An empty Path records that only run.sh starts the server, so a later
	// reinstall knows there is no old jar to clean up.
	state.ServerFile = &packstate.ServerFile{}
	var full string
	if file != "" {
		full, err = files.InDir(serverDir, file)
		if err != nil {
			return err
		}
		h, err := hashFile(full)
		if err != nil {
			return err
		}
		if h == nil {
			return fmt.Errorf("server installer reported %s, which does not exist", file)
		}
		state.ServerFile = &packstate.ServerFile{Path: file, Hashes: *h}
	}

	switch {
	case recordedPath != "" && recorded.Path != file && !(full != "" && sameFile(recordedPath, full)):
		err = os.Remove(recordedPath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("deleting old server file %s: %w", recorded.Path, err)
		}
		if err == nil {
			log.Printf("Deleted old server file: %s\n", recorded.Path)
		}
	case recorded == nil && old != nil:
		log.Printf("No server file was recorded before, so an old one may remain. Start the server with: %s\n", startWith(file))
	}
	return nil
}

// backUpUnless moves the file at path into backups, unless it is missing or
// matches h. A nil h matches nothing.
func backUpUnless(serverDir string, rel string, path string, h *modrinth.Hashes, backups *backup.Dir) error {
	disk, err := hashFile(path)
	if err != nil || disk == nil {
		return err
	}
	if h != nil && matches(*disk, *h) {
		return nil
	}
	err = backups.Move(serverDir, rel)
	if err != nil {
		return fmt.Errorf("backing up %s: %w", rel, err)
	}
	return nil
}

func startWith(file string) string {
	if file == "" {
		return "run.sh"
	}
	return file
}
