// Package cache keeps downloaded pack archives outside the server dir,
// named by their SHA-512, so a later run for the same archive reuses them.
package cache

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Dir returns the archive cache dir: mrpack-install in the user cache dir,
// created when missing. When that is unavailable or not writable (no HOME,
// a read-only root filesystem), it returns a new private dir in the OS temp
// dir instead. A shared, predictable temp path would let another user swap
// an archive between its hash check and its use, so nothing is reused then.
func Dir() (string, error) {
	return dir(os.UserCacheDir, os.TempDir())
}

func dir(userCacheDir func() (string, error), tempDir string) (string, error) {
	user, err := userCacheDir()
	if err == nil {
		d := filepath.Join(user, "mrpack-install")
		err = writableDir(d)
		if err == nil {
			return d, nil
		}
	}
	reason := err
	d, err := os.MkdirTemp(tempDir, "mrpack-install-")
	if err != nil {
		return "", err
	}
	log.Printf("Caching archives in %s for this run only, the user cache dir is not usable: %s\n", d, reason)
	return d, nil
}

func writableDir(d string) error {
	err := os.MkdirAll(d, 0755)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(d, ".probe-*")
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(f.Name())
}

// Archive returns the path of the archive at url in the cache dir d,
// downloading it with download unless a copy matching sum is cached. An
// empty sum means the sum is unknown, so the archive is always
// downloaded. The download goes to a temp file and is renamed to its sum
// once complete, so concurrent runs never see a partial archive.
func Archive(d string, url string, sum string, download func(url string, w io.Writer) error) (string, error) {
	sum = strings.ToLower(sum)
	if sum != "" && !sha512Hex.MatchString(sum) {
		return "", fmt.Errorf("downloading %s: %q is not a SHA-512", url, sum)
	}
	err := os.MkdirAll(d, 0755)
	if err != nil {
		return "", err
	}
	if sum != "" {
		path := filepath.Join(d, sum+".mrpack")
		got, err := sumFile(path)
		if err == nil && got == sum {
			log.Printf("Using cached archive %s\n", path)
			return path, nil
		}
	}

	tmp, err := os.CreateTemp(d, "*.mrpack.tmp")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	h := sha512.New()
	err = download(url, io.MultiWriter(tmp, h))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", url, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if sum != "" && got != sum {
		return "", fmt.Errorf("downloading %s: SHA-512 is %s, want %s", url, got, sum)
	}
	path := filepath.Join(d, got+".mrpack")
	err = os.Rename(tmp.Name(), path)
	if err != nil {
		// On Windows the rename fails while another run has the same
		// archive open. Its copy is just as good.
		if cached, serr := sumFile(path); serr == nil && cached == got {
			return path, nil
		}
		return "", err
	}
	return path, nil
}

var sha512Hex = regexp.MustCompile(`^[0-9a-f]{128}$`)

func sumFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha512.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
