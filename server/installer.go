package server

import (
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"log"
)

type Installer interface {
	Install(serverDir string, serverFile string) error
}

func InstallerFromDeps(deps *mrpack.Dependencies) Installer {
	var flavor Flavor
	if deps.Fabric != "" {
		flavor = Fabric
	} else if deps.Quilt != "" {
		flavor = Quilt
	} else if deps.Forge != "" {
		flavor = Forge
	} else {
		flavor = Vanilla
	}
	inst, err := NewInstaller(flavor, deps.Minecraft, "")
	if err != nil {
		log.Fatalln(err)
	}
	return inst
}

func NewInstaller(flavor Flavor, minecraftVersion string, flavorVersion string) (Installer, error) {
	var inst Installer = nil
	switch flavor {
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
	case Paper:
		inst = &PaperInstaller{
			MinecraftVersion: minecraftVersion,
		}
	default:
		log.Fatalln("No installer for flavor: " + flavor)
	}
	return inst, nil
}
