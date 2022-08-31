package server

import (
	"log"
	"net/url"
)

type VanillaSupplier struct {
	MinecraftVersion string
}

func (*VanillaSupplier) get(mcVer string, loaderVer string) (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
