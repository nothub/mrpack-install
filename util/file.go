package util

import "os"

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
