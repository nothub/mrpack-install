package server

import (
	"log"
	"net/url"
)

type QuiltSupplier struct {
	MinecraftVersion string
	QuiltVersion     string
}

func (*QuiltSupplier) GetUrl() (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
