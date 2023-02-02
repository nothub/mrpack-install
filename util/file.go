package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DetectType int8

const (
	PathMatchHashMatch   DetectType = 0
	PathMatchHashNoMatch DetectType = 1
	PathNoMatch          DetectType = 2
)

func PathIsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func PathIsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsDir()
}

func ResolvePath(path string) (string, error) {
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

func PathIsSubpath(subPath string, basePath string) (bool, error) {
	subPath, err := ResolvePath(subPath)
	if err != nil {
		return false, err
	}

	basePath, err = ResolvePath(basePath)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(subPath, basePath), nil
}

func FileDetection(hash string, path string) DetectType {
	_, err := os.Stat(path)
	if err != nil {
		return PathNoMatch
	}
	if tmp, _ := CheckFileSha1(hash, path); tmp {
		return PathMatchHashMatch
	} else {
		return PathMatchHashNoMatch
	}
}

func RemoveEmptyDir(dir string) {
	fileNames := make([]string, 0)
	dirNames := make([]string, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirNames = append(dirNames, path)
		} else {
			fileNames = append(fileNames, path)
		}
		return err
	})
	if err != nil {
		fmt.Println(err)
	}

	fileNamesAll := strings.Join(fileNames, "")

	for i := len(dirNames) - 1; i >= 0; i-- {
		if !strings.Contains(fileNamesAll, dirNames[i]) {
			err := os.Remove(dirNames[i])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
