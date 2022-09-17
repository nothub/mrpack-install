package util

import "os"

type DetectType uint8

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
