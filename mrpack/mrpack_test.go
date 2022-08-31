package mrpack

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"testing"
)

func init() {
	_, err := download("https://cdn.modrinth.com/data/1KVo5zza/versions/1vRDfe1u/MR_Fabulously%20Optimized_4.2.1.mrpack")
	if err != nil {
		log.Fatalln("download failed", err)
	}
	_, err = download("https://cdn.modrinth.com/data/KmiWHzQ4/versions/1.5.0/Skyblocker-Modpack.mrpack")
	if err != nil {
		log.Fatalln("download failed", err)
	}
}

func Test_Index_Fabulously_Optimized(t *testing.T) {
	t.Parallel()
	index, err := ReadIndex("/tmp/MR_Fabulously Optimized_4.2.1.mrpack")
	if err != nil {
		t.Fatal(err)
	}
	if index.Name != "Fabulously Optimized" {
		t.Fatal("wrong name!")
	}
}

func Test_Index_Skyblocker(t *testing.T) {
	t.Parallel()
	index, err := ReadIndex("/tmp/Skyblocker-Modpack.mrpack")
	if err != nil {
		t.Fatal(err)
	}
	if index.VersionId != "1.5.0" {
		t.Fatal("wrong name!")
	}
}

func Test_Overrides_Fabulously_Optimized(t *testing.T) {
	t.Parallel()
	err := ExtractOverrides("/tmp/MR_Fabulously Optimized_4.2.1.mrpack", "fabulously_optimized", Server)
	if err != nil {
		t.Fatal(err)
	}
	// TODO
	err = os.RemoveAll("fabulously_optimized")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Overrides_Skyblocker(t *testing.T) {
	t.Parallel()
	err := ExtractOverrides("/tmp/Skyblocker-Modpack.mrpack", "skyblocker", Client)
	if err != nil {
		t.Fatal(err)
	}
	// TODO
	err = os.RemoveAll("skyblocker")
	if err != nil {
		t.Fatal(err)
	}
}

func download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	file, err := os.Create(path.Join(os.TempDir(), path.Base(resp.Request.URL.Path)))
	if err != nil {
		return "", err
	}
	err = file.Chmod(0644)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	err = resp.Body.Close()
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}
