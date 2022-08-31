package server

import (
	"log"
	"net/url"
)

type SpigotSupplier struct {
	MinecraftVersion string
	SpigotVersion    string
}

func (*SpigotSupplier) GetUrl() (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
