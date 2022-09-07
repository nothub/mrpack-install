package server

import (
	"errors"
	"log"
	"os"
)

type Spigot struct {
	MinecraftVersion string
	SpigotVersion    string
}

func (supplier *Spigot) Provide(serverDir string, serverFile string) error {
	err := os.MkdirAll("work/spigot", 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: download https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar
	// TODO: git config --global --unset core.autocrlf
	// TODO: java -jar BuildTools.jar --rev ${minecraftVersion}

	return errors.New("spigot provider not yet implemented")
}
