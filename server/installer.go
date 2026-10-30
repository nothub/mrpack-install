package server

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Installer interface {
	// Install installs the server into serverDir and returns the jar that
	// starts it, relative to serverDir, or "" when only the run.sh and
	// run.bat the installer writes start it.
	Install(serverDir string, serverFile string) (string, error)
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

// removeInstaller deletes an installer jar and the <jar>.log it may have
// written, after a successful install, so loader upgrades don't pile them
// up in the server dir. The install worked, so a failure only warns.
func removeInstaller(jar string) {
	for _, p := range []string{jar, jar + ".log"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Unable to remove the installer: %s\n", err)
		}
	}
}

// checkJar fails unless the file at path is a jar with a manifest. Fabric's
// meta API sends no checksum for the launcher jar it builds, so this is
// what tells a download from an error page sent with status 200.
func checkJar(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("%s is not a jar: %w", filepath.Base(path), err)
	}
	defer r.Close()
	f, err := r.Open("META-INF/MANIFEST.MF")
	if err != nil {
		return fmt.Errorf("%s has no manifest: %w", filepath.Base(path), err)
	}
	return f.Close()
}
