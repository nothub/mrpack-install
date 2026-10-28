package server

import (
	"crypto"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/web"
)

// ServerStarterJar starts Forge 1.17 and later with plain java -jar, where
// the Forge installer only writes run.sh. It is LGPL-2.1:
// https://github.com/neoforged/ServerStarterJar
const (
	serverStarterJarVersion = "0.1.35"
	serverStarterJarSha256  = "d019d815868d451e57bdb965174b8443ec361336e84e9c41abc41a6c627bacd1"
)

// placeServerStarterJar downloads the pinned ServerStarterJar to
// serverDir/serverFile.
func placeServerStarterJar(serverDir string, serverFile string) error {
	url := "https://github.com/neoforged/ServerStarterJar/releases/download/" + serverStarterJarVersion + "/server.jar"
	err := downloadVerified(url, serverStarterJarSha256, serverDir, serverFile)
	if err != nil {
		return fmt.Errorf("installing ServerStarterJar: %w", err)
	}
	return nil
}

// downloadVerified downloads url to serverDir/serverFile, replacing it only
// once the download is complete and its SHA-256 is sha256Hex, so a failed
// download leaves the old file in place.
func downloadVerified(url string, sha256Hex string, serverDir string, serverFile string) error {
	tmp, err := os.CreateTemp(serverDir, "."+serverFile+".*.tmp")
	if err != nil {
		return err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	file, err := web.DefaultClient.DownloadFile(url, serverDir, filepath.Base(tmp.Name()))
	if err != nil {
		return err
	}
	ok, err := chksum.VerifyFile(file, sha256Hex, crypto.SHA256.New(), encoding.Hex)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("SHA-256 of %s does not match %s", url, sha256Hex)
	}
	return os.Rename(file, filepath.Join(serverDir, serverFile))
}
