package server

import (
	"errors"
	"strings"
)

type Provider interface {
	Provide(serverDir string, serverFile string) error
}

func NewProvider(flavor string, minecraftVersion string, flavorVersion string) (Provider, error) {
	var provider Provider = nil

	switch strings.ToLower(flavor) {
	case "vanilla":
		provider = &Vanilla{
			MinecraftVersion: minecraftVersion,
		}
	case "fabric":
		provider = &Fabric{
			MinecraftVersion: minecraftVersion,
			FabricVersion:    flavorVersion,
		}
	case "quilt":
		provider = &Quilt{
			MinecraftVersion: minecraftVersion,
			QuiltVersion:     flavorVersion,
		}
	case "forge":
		provider = &Forge{
			MinecraftVersion: minecraftVersion,
			ForgeVersion:     flavorVersion,
		}
	case "paper":
		provider = &Paper{
			MinecraftVersion: minecraftVersion,
			PaperVersion:     flavorVersion,
		}
	case "spigot":
		provider = &Spigot{
			MinecraftVersion: minecraftVersion,
			SpigotVersion:    flavorVersion,
		}
	}

	if provider == nil {
		return nil, errors.New("no provider for this flavor")
	}

	return provider, nil
}
