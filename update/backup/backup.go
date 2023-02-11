package backup

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

var dir string

func SetDir(s string) {
	if dir != "" {
		log.Fatalln("Backup directory already defined!")
	}
	dir = s
}

func Create(filePath string, serverDir string) error {
	if dir == "" {
		dir = path.Join(serverDir, "backups", time.Now().Format("2006-01-02-15-04-05"))
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
