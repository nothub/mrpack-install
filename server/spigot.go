package server

import (
	"errors"
	"os"
)

type Spigot struct {
	MinecraftVersion string
	SpigotVersion    string
}

func (provider *Spigot) Provide(serverDir string, serverFile string) error {
	return errors.New("spigot provider not yet implemented")

	err := os.MkdirAll("work/spigot", 0755)
	if err != nil {
		return err
	}

	// TODO: download https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar
	// TODO: git config --global --unset core.autocrlf
	// TODO: java -jar BuildTools.jar --rev ${minecraftVersion}

	return nil
}
