package util

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func compareSha1(verifyHash string, newDataHash string) (bool, error) {
	if newDataHash == verifyHash {
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("data Hash Error,the data sha1 is %s,but you give hash is %s", newDataHash, verifyHash))
}

func getSha1(newData *[]byte) string {
	s := sha1.New()
	s.Write(*newData)
	sha1Code := hex.EncodeToString(s.Sum(nil))
	return sha1Code
}

func CheckReadCloserSha1(verifyHash string, readCloser io.ReadCloser) (bool, error) {
	newFileHash, err := GetReadCloserSha1(readCloser)
	if err != nil {
		return false, err
	}
	return compareSha1(verifyHash, newFileHash)
}

func GetReadCloserSha1(readCloser io.ReadCloser) (string, error) {
	verifyByte, err := io.ReadAll(readCloser)
	if err != nil {
		return "", err
	}
	return getSha1(&verifyByte), nil
}

func CheckFileSha1(verifyHash string, verifyFile string) (bool, error) {
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
	return compareSha1(verifyHash, newFileHash)
}
