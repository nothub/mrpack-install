package server

import (
	"log"
	"net/url"
)

type ForgeSupplier struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (*ForgeSupplier) GetUrl() (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
