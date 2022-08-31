package server

import (
	"log"
)

type Spigot struct {
	MinecraftVersion string
	SpigotVersion    string
}

func (*Spigot) GetUrl() (string, error) {
	log.Fatalln("Not yet implemented!")
	return "", nil
}
