package backup

import (
	"log"
	"os"
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
