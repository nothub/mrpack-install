package mrpack

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
)

func Test_ParseIndex_Fabulously_Optimized(t *testing.T) {
	t.Parallel()

	zipFile := download(t, "https://cdn.modrinth.com/data/1KVo5zza/versions/1vRDfe1u/MR_Fabulously%20Optimized_4.2.1.mrpack")
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}(zipFile)

	index, err := ReadIndex(zipFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if index.Name != "Fabulously Optimized" {
		t.Fatal("wrong name!")
	}
}

func Test_ParseIndex_Skyblocker(t *testing.T) {
	t.Parallel()

	zipFile := download(t, "https://cdn.modrinth.com/data/KmiWHzQ4/versions/1.5.0/Skyblocker-Modpack.mrpack")
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}(zipFile)

	index, err := ReadIndex(zipFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if index.Name != "Skyblocker Modpack" {
		t.Fatal("wrong name!")
	}
}

func download(t *testing.T, url string) *os.File {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	nameHash := base64.StdEncoding.EncodeToString([]byte(resp.Request.URL.String()))
	pattern, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		t.Fatal(err)
	}
	nameHash = pattern.ReplaceAllString(nameHash, "")
	if len(nameHash) > 16 {
		nameHash = nameHash[len(nameHash)-16:]
	}

	file, err := os.CreateTemp(os.TempDir(), nameHash)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	return file
}
