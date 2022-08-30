package mrpack

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

func Test_ParseIndex_Mod(t *testing.T) {
	t.Parallel()
	file, err := download("https://cdn.modrinth.com/data/P7dR8mSH/versions/0.60.0+1.19.2/fabric-api-0.60.0%2B1.19.2.jar")
	if err != nil {
		t.Fatal(err)
	}
	validate(t, file)
}

func Test_ParseIndex_Pack(t *testing.T) {
	t.Parallel()
	file, err := download("https://cdn.modrinth.com/data/JE8Z4A5g/versions/ag2r5Pf8/Rinthereout-0.4.0-indev.mrpack")
	if err != nil {
		t.Fatal(err)
	}
	validate(t, file)
}

func download(url string) (*os.File, error) {
	// download from url
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// create temp file
	file, err := os.CreateTemp(os.TempDir(), base64.StdEncoding.EncodeToString([]byte(resp.Request.RequestURI)))
	if err != nil {
		return nil, err
	}

	// write body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return file, nil
}

func validate(t *testing.T, file *os.File) {
	index, err := ReadFile(file)
	if err != nil {
		return
	}
	log.Println(index.Name)

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(file.Name())
	if err != nil {
		t.Fatal(err)
	}
}
