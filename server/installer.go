package server

import (
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"log"
)

type Installer interface {
	Install(serverDir string, serverFile string) error
}

func InstallerFromDeps(deps *mrpack.Deps) Installer {
	var flavorName Flavor
	var flavorVersion string
	if deps.Fabric != "" {
		flavorName = Fabric
		flavorVersion = deps.Fabric
	} else if deps.Quilt != "" {
		flavorName = Quilt
		flavorVersion = deps.Quilt
	} else if deps.Forge != "" {
		flavorName = Forge
		flavorVersion = deps.Forge
	} else if deps.NeoForge != "" {
		flavorName = NeoForge
		flavorVersion = deps.NeoForge
	} else {
		flavorName = Vanilla
	}
	inst, err := NewInstaller(flavorName, deps.Minecraft, flavorVersion)
	if err != nil {
		log.Fatalln(err)
	}
	return inst
}

func NewInstaller(flavorName Flavor, minecraftVersion string, flavorVersion string) (Installer, error) {
	var inst Installer = nil
	switch flavorName {
	case Vanilla:
		inst = &VanillaInstaller{
			MinecraftVersion: minecraftVersion,
		}
	case Fabric:
		inst = &FabricInstaller{
			MinecraftVersion: minecraftVersion,
			FabricVersion:    flavorVersion,
		}
	case Quilt:
		inst = &QuiltInstaller{
			MinecraftVersion: minecraftVersion,
			QuiltVersion:     flavorVersion,
		}
	case Forge:
		inst = &ForgeInstaller{
			MinecraftVersion: minecraftVersion,
			ForgeVersion:     flavorVersion,
		}
	case NeoForge:
		inst = &NeoForgeInstaller{
			MinecraftVersion: minecraftVersion,
			NeoForgeVersion:  flavorVersion,
		}
	case Paper:
		inst = &PaperInstaller{
			MinecraftVersion: minecraftVersion,
		}
	default:
		log.Fatalln("No installer for flavor: " + flavorName)
	}
	return inst, nil
}
