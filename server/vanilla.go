package server

import "errors"

type Vanilla struct {
	MinecraftVersion string
}

func (supplier *Vanilla) Provide(serverDir string, serverFile string) error {
	return errors.New("vanilla provider not yet implemented")
}
