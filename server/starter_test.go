package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func serveStarter(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/server.jar"
}

func TestDownloadVerified(t *testing.T) {
	t.Parallel()
	sum := sha256.Sum256([]byte("starter"))
	dir := t.TempDir()

	err := downloadVerified(serveStarter(t, "starter"), hex.EncodeToString(sum[:]), dir, "mc.jar")
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "mc.jar"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "starter" {
		t.Errorf("mc.jar = %q, want the starter jar", got)
	}
}

// A download that fails its hash check leaves the jar it would replace.
func TestDownloadVerifiedKeepsOldFileOnWrongHash(t *testing.T) {
	t.Parallel()
	sum := sha256.Sum256([]byte("starter"))
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mc.jar"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	err := downloadVerified(serveStarter(t, "tampered"), hex.EncodeToString(sum[:]), dir, "mc.jar")

	if err == nil {
		t.Fatal("downloadVerified() = nil, want an error")
	}
	got, err := os.ReadFile(filepath.Join(dir, "mc.jar"))
	if err != nil || string(got) != "old" {
		t.Errorf("mc.jar = %q, %v, want the old file", got, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("server dir holds %v, want only mc.jar", entries)
	}
}
