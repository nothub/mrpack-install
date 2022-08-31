package server

import (
	"log"
)

type Quilt struct {
	MinecraftVersion string
	QuiltVersion     string
}

func (*Quilt) GetUrl() (string, error) {
	log.Fatalln("Not yet implemented!")
	return "", nil
}
