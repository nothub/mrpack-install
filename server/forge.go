package server

import (
	"errors"
)

type Forge struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (provider *Forge) Provide(serverDir string, serverFile string) error {
	return errors.New("forge provider not yet implemented")
}
