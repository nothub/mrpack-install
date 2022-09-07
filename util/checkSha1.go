package util

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func checkSha1(verifyHash string, verifyByte *[]byte) (bool, error) {
	s := sha1.New()
	s.Write(*verifyByte)
	sha1Code := hex.EncodeToString(s.Sum(nil))
	if sha1Code == verifyHash {
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("data Hash Error,the data sha1 is %s,but you give hash is %s", sha1Code, verifyHash))
}

func CheckResponseSha1(verifyHash string, verifyResponse *http.Response) (bool, error) {
	verifyByte, err := io.ReadAll(verifyResponse.Body)
	if err != nil {
		return false, err
	}
	err = verifyResponse.Body.Close()
	if err != nil {
		return false, err
	}
	verifyResponse.Body = io.NopCloser(verifyResponse.Body)
	return checkSha1(verifyHash, &verifyByte)
}

func CheckFileSha1(verifyHash string, verifyFile string) (bool, error) {
	file, err := os.Open(verifyFile)
	if err != nil {
		return false, err
	}
	verifyByte, err := io.ReadAll(file)
	if err != nil {
		return false, err
	}
	err = file.Close()
	if err != nil {
		return false, err
	}
	return checkSha1(verifyHash, &verifyByte)
}
