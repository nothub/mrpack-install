package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/semver"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ForgeInstaller struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (inst *ForgeInstaller) Install(serverDir string, serverFile string) (string, error) {
	// Forge 1.16.5 and older load the vanilla server jar by this name.
	if serverFile == "minecraft_server."+inst.MinecraftVersion+".jar" {
		return "", fmt.Errorf("--server-file %s would replace the vanilla jar Forge loads, choose another name", serverFile)
	}

	version := inst.ForgeVersion

	if version == "" || version == "latest" {

		meta, err := maven.FetchMetadata("https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml")
		if err != nil {
			return "", err
		}

		var versions []string
		for _, ver := range meta.Versioning.Versions {
			if !strings.HasPrefix(ver, inst.MinecraftVersion) {
				// does not match mc version
				continue
			}
			versions = append(versions, ver)
		}

		parsed, err := semver.ParseAll(versions)
		if err != nil {
			return "", err
		}

		sorted := semver.SortDesc(parsed)

		if len(sorted) <= 0 {
			return "", fmt.Errorf("no forge release version found for mc version %s", inst.MinecraftVersion)
		}

		version = sorted[0].String()
	} else {
		version = inst.MinecraftVersion + "-" + version
	}

	u := fmt.Sprintf(
		"https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar",
		version,
		version,
	)
	installerFile, err := web.DefaultClient.DownloadFile(u, ".", "")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("java", "-jar", installerFile, "--installServer", serverDir)
	log.Println("Executing command:", cmd.String())
	if err = cmd.Run(); err != nil {
		return "", err
	}

	file, err := forgeStartFile(serverDir, version, serverFile, placeServerStarterJar)
	if err != nil {
		return "", err
	}
	removeInstaller(installerFile)
	return file, nil
}

// forgeStartFile returns the jar that starts the installed Forge version.
// With serverFile set, it leaves that jar under that name: Forge 1.16.5 and
// older rename their versioned jar, and Forge 1.17 and later, which start
// through run.sh, get placeStarter's ServerStarterJar there.
func forgeStartFile(serverDir string, version string, serverFile string, placeStarter func(serverDir string, serverFile string) error) (string, error) {
	file, err := forgeServerFile(serverDir, version)
	if err != nil || serverFile == "" {
		return file, err
	}
	if file != "" && !strings.HasSuffix(file, "-shim.jar") {
		err = os.Rename(filepath.Join(serverDir, file), filepath.Join(serverDir, serverFile))
		if err != nil {
			return "", err
		}
		return serverFile, nil
	}
	err = placeStarter(serverDir, serverFile)
	if err != nil {
		return "", err
	}
	return serverFile, nil
}

// forgeServerFile returns the jar the Forge installer for version wrote to
// start the server, by its name per release range:
//   - 1.20.3 and later: forge-<version>-shim.jar, next to run.sh.
//   - 1.16.5 and older: forge-<version>.jar, or forge-<version>-universal.jar
//     on older 1.12.2 builds.
//   - 1.17 to 1.20.2: none, only run.sh starts it.
func forgeServerFile(serverDir string, version string) (string, error) {
	for _, name := range []string{
		"forge-" + version + "-shim.jar",
		"forge-" + version + ".jar",
		"forge-" + version + "-universal.jar",
	} {
		if files.IsFile(filepath.Join(serverDir, name)) {
			return name, nil
		}
	}
	if files.IsFile(filepath.Join(serverDir, "run.sh")) {
		return "", nil
	}
	return "", fmt.Errorf("forge %s install left no server jar and no run.sh", version)
}
