package server

import (
	"log"
)

type Paper struct {
	MinecraftVersion string
	PaperVersion     string
}

func (*Paper) GetUrl() (string, error) {
	log.Fatalln("Not yet implemented!")
	return "", nil
}
