package packstate

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
)

const FileName = "packstate.json"

// SchemaVersion is the schema this build reads and writes. A state without
// a schema field was written by a beta.
const SchemaVersion = 1

type Schema struct {
	SchemaVersion int    `json:"schema"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	ArchiveSha512 string `json:"archive-sha512"`
	Ids
	Options Options     `json:"options"`
	Deps    mrpack.Deps `json:"dependencies"`
	// Hashes of every pack file in place after the run, by path.
	Hashes map[string]modrinth.Hashes `json:"hashes"`
	// ServerFile is the jar the server installer produced: an empty Path
	// when only its run.sh starts the server, nil when unknown.
	ServerFile *ServerFile `json:"server-file,omitempty"`
}

// ServerFile is the main server file, relative to the server dir, with the
// hashes it had when it was installed.
type ServerFile struct {
	Path   string          `json:"path"`
	Hashes modrinth.Hashes `json:"hashes"`
}

// Options are the flags that decide which pack files a run installs.
type Options struct {
	OptionalSelect     []string `json:"optional-select"`
	OptionalDisableAll bool     `json:"optional-disable-all"`
	// Exclude holds patterns of paths no run touches, see Excludes.
	Exclude []string `json:"exclude"`
	// ServerFile is the --server-file name, "" for the installer's own.
	ServerFile string `json:"server-file"`
	// ForceSync skips the no-op check and reinstalls the server. It applies
	// to one run, so it is never recorded.
	ForceSync bool `json:"-"`
}

// Equal reports whether opts and other shape a sync the same way. The
// order of list entries doesn't matter, and a nil list equals an empty one.
func (opts Options) Equal(other Options) bool {
	return reflect.DeepEqual(opts.normalized(), other.normalized())
}

func (opts Options) normalized() Options {
	sorted := func(list []string) []string {
		c := append([]string{}, list...)
		slices.Sort(c)
		return c
	}
	opts.OptionalSelect = sorted(opts.OptionalSelect)
	opts.Exclude = sorted(opts.Exclude)
	opts.ForceSync = false
	return opts
}

// Excludes reports whether the pack-relative path matches an Exclude
// pattern: path.Match per segment, where a ** segment matches any number of
// segments, none included.
func (opts Options) Excludes(p string) bool {
	segments := strings.Split(p, "/")
	for _, pattern := range opts.Exclude {
		if matchSegments(strings.Split(pattern, "/"), segments) {
			return true
		}
	}
	return false
}

func matchSegments(pattern []string, segments []string) bool {
	if len(pattern) == 0 {
		return len(segments) == 0
	}
	if pattern[0] == "**" {
		for i := 0; i <= len(segments); i++ {
			if matchSegments(pattern[1:], segments[i:]) {
				return true
			}
		}
		return false
	}
	if len(segments) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], segments[0])
	return err == nil && ok && matchSegments(pattern[1:], segments[1:])
}

// CheckPattern fails for an Exclude pattern that could never match a pack
// path: empty, absolute, with an empty, . or .. segment, or a malformed one.
func CheckPattern(pattern string) error {
	if pattern == "" {
		return errors.New("exclude pattern is empty")
	}
	if strings.HasPrefix(pattern, "/") {
		return fmt.Errorf("exclude pattern %q is absolute, patterns are relative to the server dir", pattern)
	}
	if strings.HasSuffix(pattern, "/") {
		return fmt.Errorf("exclude pattern %q ends in /, use %q for everything below it", pattern, pattern+"**")
	}
	for _, segment := range strings.Split(pattern, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("exclude pattern %q has an empty, . or .. segment", pattern)
		}
		if _, err := path.Match(segment, ""); err != nil {
			return fmt.Errorf("exclude pattern %q: %w", pattern, err)
		}
	}
	return nil
}

// Selects reports whether an index file is installed on the server.
func (opts Options) Selects(f mrpack.File) bool {
	switch f.Env.Server {
	case modrinth.RequiredEnvSupport:
		return true
	case modrinth.OptionalEnvSupport:
		if opts.OptionalDisableAll {
			return false
		}
		if len(opts.OptionalSelect) < 1 {
			return true
		}
		for _, p := range opts.OptionalSelect {
			if regexp.MustCompile(p).Match([]byte(filepath.Base(f.Path))) {
				return true
			}
		}
		return false
	case modrinth.UnsupportedEnvSupport:
		return false
	default:
		log.Fatalf("Unexpected environment configuration %q for mod %q\n", f.Env.Server, f.Path)
		return false
	}
}

// Ids name a pack version on Modrinth.
type Ids struct {
	Slug      string `json:"slug"`
	ProjectId string `json:"project-id"`
	VersionId string `json:"version-id"`
}

// Save writes the state to a temp file and renames it into place, so a
// killed run leaves the previous state or the new one, never a torn file.
func (state *Schema) Save(serverDir string) error {
	if state == nil {
		return nil
	}

	j, err := json.Marshal(state)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, j, "", strings.Repeat(" ", 4))
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(serverDir, FileName+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(buf.Bytes())
	if err != nil {
		tmp.Close()
		return err
	}
	err = tmp.Chmod(0644)
	if err != nil {
		tmp.Close()
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(serverDir, FileName))
}

// LoadPackState returns the state in serverDir, or nil when there is none.
// A beta state, without a schema field, also counts as none.
func LoadPackState(serverDir string) (*Schema, error) {
	b, err := os.ReadFile(filepath.Join(serverDir, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state Schema
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	switch state.SchemaVersion {
	case 0:
		return nil, nil
	case SchemaVersion:
		return &state, nil
	default:
		return nil, fmt.Errorf("%s has unknown schema %d, written by a newer mrpack-install", FileName, state.SchemaVersion)
	}
}

// FromArchive builds the state an install of the *.mrpack file at path
// leaves with opts. ids looks the archive up by its SHA-512. A failed
// lookup leaves the ids empty, so a pack Modrinth doesn't know still gets
// a state.
func FromArchive(path string, opts Options, ids func(sha512 string) (Ids, error)) (*Schema, error) {
	index, err := mrpack.ReadIndex(path)
	if err != nil {
		return nil, err
	}

	archiveSha512, err := ArchiveSha512(path)
	if err != nil {
		return nil, err
	}

	var state Schema
	state.SchemaVersion = SchemaVersion
	state.Name = index.Name
	state.Version = index.Version
	state.ArchiveSha512 = archiveSha512
	state.Options = opts
	state.Deps = index.Deps
	state.Hashes = ShippedFiles(index, path, opts)

	state.Ids, err = ids(archiveSha512)
	if err != nil {
		log.Printf("Modrinth lookup failed, %s records no ids: %s\n", FileName, err)
		state.Ids = Ids{}
	}

	return &state, nil
}

// ArchiveSha512 returns the hex SHA-512 of the archive at path, the pack's
// identity for the no-op check.
func ArchiveSha512(path string) (string, error) {
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

// Invalidate saves the state without its archive hash, so a run that fails
// after this can't satisfy a later no-op check. The in-memory state keeps
// it.
func (state *Schema) Invalidate(serverDir string) error {
	if state == nil {
		return nil
	}
	c := *state
	c.ArchiveSha512 = ""
	return c.Save(serverDir)
}

// ShippedFiles returns the hashes of every file the pack at archive puts
// on the server with opts, by path. An override's content wins over the
// index file at the same path. Excluded paths are left out.
func ShippedFiles(index *mrpack.Index, archive string, opts Options) map[string]modrinth.Hashes {
	shipped := make(map[string]modrinth.Hashes)
	for _, indexFile := range index.Files {
		if !opts.Selects(indexFile) || opts.Excludes(indexFile.Path) {
			continue
		}
		shipped[indexFile.Path] = indexFile.Hashes
	}
	for p, hashes := range mrpack.OverrideHashes(archive) {
		if opts.Excludes(p) {
			continue
		}
		shipped[p] = hashes
	}
	return shipped
}
