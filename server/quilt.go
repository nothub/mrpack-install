package server

import (
	"errors"
	"os"
)

type Quilt struct {
	MinecraftVersion string
	QuiltVersion     string
}

func (provider *Quilt) Provide(serverDir string, serverFile string) error {
	return errors.New("quilt provider not yet implemented")

	err := os.MkdirAll("work/quilt", 0755)
	if err != nil {
		return err
	}

	// TODO: download https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/latest/quilt-installer-latest.jar
	// TODO: java -jar quilt-installer-latest.jar install server ${minecraftVersion} --download-server

	return nil
}
