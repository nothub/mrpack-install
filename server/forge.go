package server

import (
	"errors"
	"fmt"
)

type ForgeInstaller struct {
	MinecraftVersion string
	ForgeVersion     string
}

// TODO maven version lookup: https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml

func (inst *ForgeInstaller) Install(serverDir string, serverFile string) error {
	u := "https://files.minecraftforge.net/net/minecraftforge/forge/index_" + inst.MinecraftVersion + ".html"
	fmt.Println("Please acquire the required forge server file ("+inst.ForgeVersion+") manually to continue:", u)
	return errors.New("forge provider not implemented")
}
