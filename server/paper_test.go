package server

import (
	"regexp"
	"testing"
)

const paperUrlPattern string = "https://api\\.papermc\\.io/v2/projects/paper/versions/\\d+\\.\\d+\\.\\d+/builds/\\d+/downloads/paper-\\d+\\.\\d+\\.\\d+-\\d+\\.jar"

func Test_Paper_latest(t *testing.T) {
	t.Parallel()

	var supplier DownloadSupplier
	supplier = &Paper{
		MinecraftVersion: "1.18.2",
		PaperVersion:     "latest",
	}
	url, err := supplier.GetUrl()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(paperUrlPattern).MatchString(url) {
		t.Fatal("wrong url!")
	}
}

func Test_Paper_specific(t *testing.T) {
	t.Parallel()
	var supplier DownloadSupplier
	supplier = &Paper{
		MinecraftVersion: "1.12.2",
		PaperVersion:     "1619",
	}
	url, err := supplier.GetUrl()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(paperUrlPattern).MatchString(url) {
		t.Fatal("wrong url!")
	}
}

func Test_Paper_Empty_Version(t *testing.T) {
	t.Parallel()
	var supplier DownloadSupplier
	supplier = &Paper{
		MinecraftVersion: "1.9001.42",
	}
	_, err := supplier.GetUrl()
	if err.Error() != "http status 404" {
		t.Fatal("wrong status!")
	}
}
