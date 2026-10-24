package backup

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nothub/mrpack-install/files"
)

var dir string

func SetDir(s string) {
	if dir != "" {
		log.Fatalln("Backup directory already defined!")
	}
	dir = s
}

func Create(filePath string, serverDir string) error {
	log.Printf("Backup: %s\n", filePath)

	if dir == "" {
		dir = filepath.Join(serverDir, "backups", time.Now().Format("20060102150405"))
	}

	// create backup dirs
	err := os.MkdirAll(filepath.Dir(filepath.Join(dir, filePath)), 0755)
	if err != nil {
		return err
	}

	// move file to backups
	err = os.Rename(filepath.Join(serverDir, filePath), filepath.Join(dir, filePath))
	if err != nil {
		return err
	}

	return nil
}

// Dir is one run's backup dir, <base>/<timestamp>/, created on the first
// Move so a run that backs up nothing leaves no trace.
type Dir struct {
	base string
	path string
}

func NewDir(base string) *Dir {
	return &Dir{base: base}
}

// Move moves serverDir/rel into the run dir under the same relative path.
func (d *Dir) Move(serverDir string, rel string) error {
	if d.path == "" {
		path, err := createRunDir(d.base)
		if err != nil {
			return err
		}
		d.path = path
	}
	src, err := files.InDir(serverDir, rel)
	if err != nil {
		return err
	}
	log.Printf("Backup: %s\n", rel)
	return files.Move(src, filepath.Join(d.path, rel))
}

// createRunDir creates a run dir under base that no earlier run used,
// suffixing the timestamp when two runs start in the same second.
func createRunDir(base string) (string, error) {
	err := os.MkdirAll(base, 0755)
	if err != nil {
		return "", err
	}
	name := time.Now().Format("20060102150405")
	for i := 0; ; i++ {
		path := filepath.Join(base, name)
		if i > 0 {
			path = fmt.Sprintf("%s-%d", path, i)
		}
		err := os.Mkdir(path, 0755)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		return path, nil
	}
}
