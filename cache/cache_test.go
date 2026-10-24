package cache

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const data = "archive"

func sum(s string) string {
	h := sha512.Sum512([]byte(s))
	return hex.EncodeToString(h[:])
}

// serve returns a download func that writes content and counts its calls.
func serve(content string, calls *int) func(string, io.Writer) error {
	return func(_ string, w io.Writer) error {
		*calls++
		_, err := io.WriteString(w, content)
		return err
	}
}

func assertOnly(t *testing.T, dir string, name string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != name {
		t.Errorf("cache dir holds %v, want only %s", entries, name)
	}
}

func TestArchiveDownloadsUnknownSumAndNamesItBySum(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	calls := 0

	path, err := Archive(dir, "https://example.org/pack.mrpack", "", serve(data, &calls))
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(dir, sum(data)+".mrpack"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if calls != 1 {
		t.Errorf("downloads = %d, want 1", calls)
	}
	assertOnly(t, dir, sum(data)+".mrpack")
}

func TestArchiveAcceptsUppercaseSum(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cached := filepath.Join(dir, sum(data)+".mrpack")
	if err := os.WriteFile(cached, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	calls := 0

	path, err := Archive(dir, "https://example.org/pack.mrpack", strings.ToUpper(sum(data)), serve(data, &calls))
	if err != nil {
		t.Fatal(err)
	}

	if path != cached || calls != 0 {
		t.Errorf("path, downloads = %q, %d, want %q, 0", path, calls, cached)
	}
}

func TestArchiveRejectsMalformedSum(t *testing.T) {
	t.Parallel()
	calls := 0

	path, err := Archive(t.TempDir(), "https://example.org/pack.mrpack", "../../evil", serve(data, &calls))

	if err == nil {
		t.Errorf("Archive() = %q, want an error", path)
	}
}

func TestArchiveReusesValidCachedArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cached := filepath.Join(dir, sum(data)+".mrpack")
	if err := os.WriteFile(cached, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	calls := 0

	path, err := Archive(dir, "https://example.org/pack.mrpack", sum(data), serve(data, &calls))
	if err != nil {
		t.Fatal(err)
	}

	if path != cached {
		t.Errorf("path = %q, want %q", path, cached)
	}
	if calls != 0 {
		t.Errorf("downloads = %d, want 0", calls)
	}
}

func TestArchiveReplacesTruncatedCachedArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cached := filepath.Join(dir, sum(data)+".mrpack")
	if err := os.WriteFile(cached, []byte(data[:3]), 0644); err != nil {
		t.Fatal(err)
	}
	calls := 0

	path, err := Archive(dir, "https://example.org/pack.mrpack", sum(data), serve(data, &calls))
	if err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Errorf("downloads = %d, want 1", calls)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != data {
		t.Errorf("cached = %q, want %q", got, data)
	}
	assertOnly(t, dir, sum(data)+".mrpack")
}

func TestArchiveRejectsDownloadWithWrongSum(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	calls := 0

	path, err := Archive(dir, "https://example.org/pack.mrpack", sum(data), serve("tampered", &calls))

	if err == nil {
		t.Fatalf("Archive() = %q, want an error", path)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("cache dir holds %v, want nothing", entries)
	}
}

func TestArchiveRemovesTempFileOnFailedDownload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	failing := func(_ string, w io.Writer) error {
		_, _ = io.WriteString(w, "part")
		return errors.New("connection reset")
	}

	if _, err := Archive(dir, "https://example.org/pack.mrpack", "", failing); err == nil {
		t.Fatal("Archive() error = nil, want an error")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("cache dir holds %v, want nothing", entries)
	}
}

func TestDirUsesUserCacheDir(t *testing.T) {
	t.Parallel()
	user := t.TempDir()
	temp := t.TempDir()

	got, err := dir(func() (string, error) { return user, nil }, temp)
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(user, "mrpack-install"); got != want {
		t.Errorf("dir = %q, want %q", got, want)
	}
}

func TestDirFallsBackToTempDir(t *testing.T) {
	t.Parallel()
	// A regular file where the user cache dir should be makes it impossible
	// to create.
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0644); err != nil {
		t.Fatal(err)
	}
	readOnly := t.TempDir()
	if err := os.Mkdir(filepath.Join(readOnly, "mrpack-install"), 0555); err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	tests := map[string]func() (string, error){
		"unavailable":   func() (string, error) { return "", errors.New("$HOME is not defined") },
		"cannot create": func() (string, error) { return blocked, nil },
	}
	// root writes into a 0555 dir anyway.
	if os.Geteuid() != 0 {
		tests["not writable"] = func() (string, error) { return readOnly, nil }
	}
	for name, user := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := dir(user, temp)
			if err != nil {
				t.Fatal(err)
			}

			if filepath.Dir(got) != temp || !strings.HasPrefix(filepath.Base(got), "mrpack-install-") {
				t.Errorf("dir = %q, want a mrpack-install-* dir in %q", got, temp)
			}
			fi, err := os.Stat(got)
			if err != nil {
				t.Fatal(err)
			}
			if fi.Mode().Perm() != 0700 {
				t.Errorf("dir mode = %v, want 0700", fi.Mode().Perm())
			}
			if again, _ := dir(user, temp); again == got {
				t.Errorf("second dir = %q, want a new one", again)
			}
		})
	}
}
