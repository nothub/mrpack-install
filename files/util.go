package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IsFile reports whether path is something other than a dir. A missing
// path is no file, other stat errors are returned.
func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// InDir joins the pack-relative path rel to dir. It fails when rel is
// absolute, names dir itself, or lies outside dir once symlinks resolve,
// including symlinks in the part of the path that already exists.
func InDir(dir string, rel string) (string, error) {
	if filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" || filepath.Clean(rel) == "." {
		return "", fmt.Errorf("file path is not safe: %q", rel)
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	joined := filepath.Join(dir, rel)

	base, err := resolveExisting(dir)
	if err != nil {
		return "", fmt.Errorf("checking file path %q: %w", rel, err)
	}
	target, err := resolveExisting(joined)
	if err != nil {
		return "", fmt.Errorf("checking file path %q: %w", rel, err)
	}
	r, err := filepath.Rel(base, target)
	if err != nil || r == "." || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("file path is not safe: %q", rel)
	}
	return joined, nil
}

// resolveExisting resolves the symlinks of the longest existing prefix of
// the absolute path and appends the rest. A dangling symlink fails, since
// writing through it would create its target.
func resolveExisting(path string) (string, error) {
	rest := ""
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			return filepath.Join(resolved, rest), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if _, err := os.Lstat(path); err == nil {
			return "", fmt.Errorf("dangling symlink %s", path)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return filepath.Join(path, rest), nil
		}
		rest = filepath.Join(filepath.Base(path), rest)
		path = parent
	}
}

// rename is swapped by tests to simulate a cross-device move.
var rename = os.Rename

// Move moves the file at src to dst, creating dst's parent dirs. It fails
// when dst exists. A move across filesystems, such as into a bind-mounted
// dir, falls back to copy and delete.
func Move(src string, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(dst); err == nil {
		return fmt.Errorf("moving %s: %s exists", src, dst)
	}
	err = rename(src, dst)
	if !errors.Is(err, errCrossDevice) {
		return err
	}
	err = copyFile(src, dst)
	if err != nil {
		return fmt.Errorf("copying %s across filesystems: %w", src, err)
	}
	return os.Remove(src)
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	fi, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fi.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err == nil {
		// The source is removed next, so the copy must be on disk first.
		err = out.Sync()
	}
	if err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	err = out.Close()
	if err != nil {
		os.Remove(dst)
		return err
	}
	return os.Chmod(dst, fi.Mode().Perm())
}
