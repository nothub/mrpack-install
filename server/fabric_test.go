package server

import (
	"regexp"
	"testing"
)

const fabricUrlPattern string = "https://meta\\.fabricmc\\.net/v2/versions/loader/\\d+\\.\\d+\\.\\d+/\\d+\\.\\d+\\.\\d+/\\d+\\.\\d+\\.\\d+/server/jar"

func Test_Fabric_latest(t *testing.T) {
	t.Parallel()

	var supplier DownloadSupplier
	supplier = &Fabric{
		MinecraftVersion: "1.18.2",
		FabricVersion:    "latest",
	}
	url, err := supplier.GetUrl()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(fabricUrlPattern).MatchString(url) {
		t.Fatal("wrong url!")
	}
}

func Test_Fabric_specific(t *testing.T) {
	t.Parallel()
	var supplier DownloadSupplier
	supplier = &Fabric{
		MinecraftVersion: "1.19.2",
		FabricVersion:    "0.14.9",
	}
	url, err := supplier.GetUrl()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(fabricUrlPattern).MatchString(url) {
		t.Fatal("wrong url!")
	}
}

func Test_Fabric_Empty_Version(t *testing.T) {
	t.Parallel()
	var supplier DownloadSupplier
	supplier = &Fabric{
		MinecraftVersion: "1.16.5",
	}
	url, err := supplier.GetUrl()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(fabricUrlPattern).MatchString(url) {
		t.Fatal("wrong url!")
	}
}
