package server

import (
	"log"
	"net/url"
)

type PaperSupplier struct {
	MinecraftVersion string
	PaperVersion     string
}

func (*PaperSupplier) GetUrl() (*url.URL, error) {
	log.Fatalln("Not yet implemented!")
	return nil, nil
}
