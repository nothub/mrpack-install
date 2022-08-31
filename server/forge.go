package server

import (
	"log"
)

type Forge struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (*Forge) GetUrl() (string, error) {
	log.Fatalln("Not yet implemented!")
	return "", nil
}
