package server

import (
	"log"
	"net/url"
)

type ForgeSupplier struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (*ForgeSupplier) get(mcVer string, loaderVer string) (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
