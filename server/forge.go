package server

import (
	"errors"
	"fmt"
)

type Forge struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (provider *Forge) Provide(serverDir string, serverFile string) error {
	u := "https://files.minecraftforge.net/net/minecraftforge/forge/index_" + provider.MinecraftVersion + ".html"
	fmt.Println("Please acquire the required forge server file ("+provider.ForgeVersion+") manually to continue:", u)
	return errors.New("forge provider not implemented")
}
