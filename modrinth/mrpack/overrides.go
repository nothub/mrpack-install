package mrpack

import (
	"archive/zip"
	"crypto"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/files"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const pathPrefixBoth = "overrides/"
const pathPrefixServer = "server-overrides/"
const pathPrefixClient = "client-overrides/"

type override zip.File

func (o *override) prefix() string {
	if strings.HasPrefix(o.Name, pathPrefixBoth) {
		return pathPrefixBoth
	}
	if strings.HasPrefix(o.Name, pathPrefixServer) {
		return pathPrefixServer
	}
	if strings.HasPrefix(o.Name, pathPrefixClient) {
		return pathPrefixClient
	}
	panic("Not an override path: " + o.Name)
}

func (o *override) realPath() string {
	return strings.TrimPrefix(o.Name, o.prefix())
}

func (o *override) server() bool {
	return strings.HasPrefix(o.Name, pathPrefixBoth) ||
		strings.HasPrefix(o.Name, pathPrefixServer)
}

func ExtractOverrides(zipFile string, serverDir string) (err error) {
	return IterZip(zipFile, func(file *zip.File) error {
		o := override(*file)
		if !o.server() {
			// skip non-server override files
			return nil
		}
		p := o.realPath()
		files.AssertSafe(p, serverDir)
		targetPath := filepath.Join(serverDir, p)

		err := os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			return err
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, fileReader)
		if err != nil {
			return err
		}

		err = fileReader.Close()
		if err != nil {
			return err
		}
		err = outFile.Close()
		if err != nil {
			return err
		}

		log.Printf("Override: %s\n", targetPath)

		return nil
	})
}

func OverrideHashes(zipFile string) map[string]modrinth.Hashes {
	overrides := make(map[string]modrinth.Hashes)

	err := IterZip(zipFile, func(file *zip.File) error {
		o := override(*file)
		if !o.server() {
			// skip non-server override files
			return nil
		}
		p := o.realPath()

		r, err := file.Open()
		if err != nil {
			return err
		}

		var hashes modrinth.Hashes
		h, err := chksum.Create(r, crypto.SHA1.New(), encoding.Hex)
		if err != nil {
			return err
		}
		hashes.Sha1 = h
		h, err = chksum.Create(r, crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			return err
		}
		hashes.Sha512 = h
		overrides[p] = hashes

		err = r.Close()
		if err != nil {
			log.Println(err.Error())
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	return overrides
}
