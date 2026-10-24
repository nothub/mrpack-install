package download

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
)

const jarData = "jar"

func jarHashes() modrinth.Hashes {
	sum := sha512.Sum512([]byte(jarData))
	return modrinth.Hashes{Sha512: hex.EncodeToString(sum[:])}
}

// serve answers /ok.jar with jarData and every other path with 404, and
// counts the requests it got.
func serve(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/ok.jar" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(jarData))
	}))
	t.Cleanup(srv.Close)
	return srv, &requests
}

func TestDownloadReportsFailedFiles(t *testing.T) {
	t.Parallel()
	srv, _ := serve(t)
	dir := t.TempDir()
	d := Downloader{
		Downloads: []*Download{
			{Path: "mods/ok.jar", Urls: []string{srv.URL + "/ok.jar"}, Hashes: jarHashes()},
			{Path: "mods/missing.jar", Urls: []string{srv.URL + "/missing.jar"}, Hashes: jarHashes()},
		},
		Threads: 2,
		Retries: 2,
	}

	err := d.Download(dir)

	if err == nil {
		t.Fatal("err = nil, want an error naming mods/missing.jar")
	}
	if !strings.Contains(err.Error(), "mods/missing.jar") {
		t.Errorf("err = %q, want it to name mods/missing.jar", err)
	}
	if strings.Contains(err.Error(), "mods/ok.jar") {
		t.Errorf("err = %q, names mods/ok.jar, which downloaded", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mods", "ok.jar")); err != nil {
		t.Error(err)
	}
}

func TestDownloadWithZeroRetriesTriesOnce(t *testing.T) {
	t.Parallel()
	srv, requests := serve(t)
	d := Downloader{
		Downloads: []*Download{
			{Path: "mods/missing.jar", Urls: []string{srv.URL + "/missing.jar"}, Hashes: jarHashes()},
		},
		Threads: 1,
		Retries: 0,
	}

	if err := d.Download(t.TempDir()); err == nil {
		t.Fatal("err = nil, want an error")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestDownloadRetriesCountsAttempts(t *testing.T) {
	t.Parallel()
	srv, requests := serve(t)
	d := Downloader{
		Downloads: []*Download{
			{Path: "mods/missing.jar", Urls: []string{srv.URL + "/missing.jar"}, Hashes: jarHashes()},
		},
		Threads: 1,
		Retries: 3,
	}

	if err := d.Download(t.TempDir()); err == nil {
		t.Fatal("err = nil, want an error")
	}
	if got := requests.Load(); got != 3 {
		t.Errorf("requests = %d, want 3", got)
	}
}

func TestDownloadWithoutUrlsFails(t *testing.T) {
	t.Parallel()
	d := Downloader{
		Downloads: []*Download{{Path: "mods/nourl.jar", Hashes: jarHashes()}},
		Threads:   1,
		Retries:   3,
	}

	err := d.Download(t.TempDir())

	if err == nil || !strings.Contains(err.Error(), "mods/nourl.jar") {
		t.Errorf("err = %v, want an error naming mods/nourl.jar", err)
	}
}
