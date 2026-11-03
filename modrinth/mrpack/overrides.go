package mrpack

import (
	"archive/zip"
	"bytes"
	"crypto"
	"fmt"
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

// iterServerOverrides visits every overrides/ entry before any
// server-overrides/ entry, whatever the zip order, so a path present in both
// ends up with the server-overrides copy.
func iterServerOverrides(zipFile string, callback func(file *zip.File) error) error {
	for _, prefix := range []string{pathPrefixBoth, pathPrefixServer} {
		// Two entries for one path in one layer leave which content is meant
		// to the zip's entry order, so such a pack is rejected.
		seen := make(map[string]bool)
		err := IterZip(zipFile, func(file *zip.File) error {
			if !strings.HasPrefix(file.Name, prefix) {
				return nil
			}
			if seen[file.Name] {
				return fmt.Errorf("pack has %s more than once", file.Name)
			}
			seen[file.Name] = true
			return callback(file)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ExtractOverrides writes the server's override files to serverDir. With a
// non-nil keep, only the paths keep accepts are written.
func ExtractOverrides(zipFile string, serverDir string, keep func(path string) bool) (err error) {
	return iterServerOverrides(zipFile, func(file *zip.File) error {
		o := override(*file)
		p := o.realPath()
		if keep != nil && !keep(p) {
			return nil
		}
		targetPath, err := files.InDir(serverDir, p)
		if err != nil {
			return err
		}

		err = writeOverride(file, targetPath)
		if err != nil {
			return err
		}
		log.Printf("Override: %s\n", targetPath)
		return nil
	})
}

// writeOverride writes the content of the zip entry to target, creating
// its parent dirs.
func writeOverride(file *zip.File, target string) error {
	err := os.MkdirAll(filepath.Dir(target), 0755)
	if err != nil {
		return err
	}
	in, err := file.Open()
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

func OverrideHashes(zipFile string) (map[string]modrinth.Hashes, error) {
	overrides := make(map[string]modrinth.Hashes)

	err := iterServerOverrides(zipFile, func(file *zip.File) error {
		o := override(*file)
		p := o.realPath()

		r, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(r)
		if cerr := r.Close(); cerr != nil {
			log.Println(cerr.Error())
		}
		if err != nil {
			return err
		}

		var hashes modrinth.Hashes
		h, err := chksum.Create(bytes.NewReader(data), crypto.SHA1.New(), encoding.Hex)
		if err != nil {
			return err
		}
		hashes.Sha1 = h
		h, err = chksum.Create(bytes.NewReader(data), crypto.SHA512.New(), encoding.Hex)
		if err != nil {
			return err
		}
		hashes.Sha512 = h
		overrides[p] = hashes

		return nil
	})
	if err != nil {
		return nil, err
	}

	return overrides, nil
}
