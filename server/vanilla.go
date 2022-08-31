package server

import (
	"log"
)

type Vanilla struct {
	MinecraftVersion string
}

func (*Vanilla) GetUrl() (string, error) {
	log.Fatalln("Not yet implemented!")
	return "", nil
}
