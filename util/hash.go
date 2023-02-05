package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// Hashes maps file paths to hashes.
type Hashes map[string]string

func Sha1(data *[]byte) string {
	s := sha1.New()
	s.Write(*data)
	return hex.EncodeToString(s.Sum(nil))
}

func GetReadCloserSha1(readCloser io.ReadCloser) (string, error) {
	verifyByte, err := io.ReadAll(readCloser)
	if err != nil {
		return "", err
	}
	return Sha1(&verifyByte), nil
}

func CheckFileSha1(verifyHash string, verifyFile string) (bool, error) {
	_, err := os.Stat(verifyFile)
	if err != nil {
		log.Fatalln("The validated file does not exist", verifyFile)
	}
	file, err := os.Open(verifyFile)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err != nil {
		return false, err
	}
	newFileHash, err := GetReadCloserSha1(file)
	if err != nil {
		return false, err
	}
	return verifyHash == newFileHash, nil
}
