package files

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		log.Fatalln(err.Error())
	}
	return !info.IsDir()
}

func Resolve(path string) (string, error) {
	// resolve absolute path
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); err == nil {
		// resolve symlinks
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return "", err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	return path, nil
}

func IsSubpath(subPath string, basePath string) (bool, error) {
	subPath, err := Resolve(subPath)
	if err != nil {
		return false, err
	}

	basePath, err = Resolve(basePath)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(subPath, basePath), nil
}

func AssertSafe(subPath string, basePath string) {
	ok, err := IsSubpath(subPath, basePath)
	if err != nil {
		log.Println(err.Error())
	}
	if err != nil || !ok {
		log.Fatalln("File path is not safe: " + subPath)
	}
}

func CountFiles(dir string) int {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			count = count + 1
		}
		return err
	})
	if err != nil {
		log.Fatalln("Unable to walk file tree!", err.Error())
	}
	return count
}

func RmEmptyDirs(dir string) {
	var dirs []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// ignore tree root
		if dir == path {
			return err
		}
		if info.IsDir() && CountFiles(path) == 0 {
			// prepend because we want to delete the innermost children first
			dirs = append([]string{path}, dirs...)
		}
		return err
	})
	if err != nil {
		log.Fatalln("Unable to walk file tree!", err.Error())
	}

	for _, path := range dirs {
		err := os.Remove(path)
		if err != nil {
			log.Printf("Unable to delete empty directory %s. %s\n", path, err.Error())
		}
	}
}
