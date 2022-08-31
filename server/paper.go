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
	// https://api.papermc.io/v2/projects/paper/versions/1.19.2/builds
	// -> builds[] last element -> latest build -> downloads -> application -> name
	// https://api.papermc.io/v2/projects/paper/versions/1.19.2/builds/BUILD_ID/downloads/NAME
	return "", nil
}
