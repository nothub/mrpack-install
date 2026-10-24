package server

import (
	"fmt"
	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/maven"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/semver"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type ForgeInstaller struct {
	MinecraftVersion string
	ForgeVersion     string
}

func (inst *ForgeInstaller) Install(serverDir string, serverFile string) (string, error) {

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

	if serverFile != "" {
		log.Println("ignoring --server-file option for forge server installation!")
	}

	return forgeServerFile(serverDir, version)
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
