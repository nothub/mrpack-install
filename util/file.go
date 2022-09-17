package util

import (
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
		panic(err)
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
