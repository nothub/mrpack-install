package server

import (
	"log"
	"net/url"
)

type SpigotSupplier struct {
	MinecraftVersion string
	SpigotVersion    string
}

func (*SpigotSupplier) get(mcVer string, loaderVer string) (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
