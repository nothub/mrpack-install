package packstate

import (
	"bytes"
	"crypto/sha1"
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
// A beta state, without a schema field, is converted to the current schema.
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
		return convertBeta(serverDir, b)
	case SchemaVersion:
		return &state, nil
	default:
		return nil, fmt.Errorf("%s has unknown schema %d, written by a newer mrpack-install", FileName, state.SchemaVersion)
	}
}

// emptySha512 is the SHA-512 of empty input. Betas recorded it for every
// override, because the SHA-1 pass drained the entry first.
const emptySha512 = "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"

// betaSchema is packstate.json as v0.17.0-beta to v0.21.0-beta wrote it.
type betaSchema struct {
	Slug      string                     `json:"slug"`
	ProjectId string                     `json:"project-id"`
	Version   string                     `json:"version"`
	VersionId string                     `json:"version-id"`
	Deps      mrpack.Deps                `json:"dependencies"`
	Hashes    map[string]modrinth.Hashes `json:"hashes"`
}

// convertBeta reads a beta state as the current schema. It has no archive
// hash, so it never satisfies a no-op, and its server file is inferred.
func convertBeta(serverDir string, b []byte) (*Schema, error) {
	var beta betaSchema
	err := json.Unmarshal(b, &beta)
	if err != nil {
		return nil, err
	}
	state := Schema{
		SchemaVersion: SchemaVersion,
		Version:       beta.Version,
		Ids:           Ids{Slug: beta.Slug, ProjectId: beta.ProjectId, VersionId: beta.VersionId},
		Deps:          beta.Deps,
		Hashes:        make(map[string]modrinth.Hashes, len(beta.Hashes)),
	}
	for p, h := range beta.Hashes {
		if h.Sha512 == emptySha512 {
			h.Sha512 = ""
		}
		state.Hashes[p] = h
	}
	state.ServerFile, err = inferServerFile(serverDir, beta.Deps)
	if err != nil {
		return nil, err
	}
	log.Printf("Reading the beta %s as schema %d\n", FileName, SchemaVersion)
	return &state, nil
}

// inferServerFile finds the server jar an old state never recorded, under
// a name derived from its own dependencies: Forge's versioned jar by exact
// name, or Fabric's launcher with the launcher version as the only
// wildcard. Only a single match is recorded, with its current hashes.
func inferServerFile(serverDir string, deps mrpack.Deps) (*ServerFile, error) {
	var matches []string
	switch {
	case deps.Forge != "":
		version := deps.Minecraft + "-" + deps.Forge
		for _, name := range []string{
			"forge-" + version + ".jar",
			"forge-" + version + "-universal.jar",
			"forge-" + version + "-shim.jar",
		} {
			if fi, err := os.Lstat(filepath.Join(serverDir, name)); err == nil && fi.Mode().IsRegular() {
				matches = append(matches, name)
			}
		}
	case deps.Fabric != "":
		pattern := "fabric-server-mc." + escapeGlob(deps.Minecraft) + "-loader." + escapeGlob(deps.Fabric) + "-launcher.*.jar"
		entries, err := os.ReadDir(serverDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if ok, _ := path.Match(pattern, e.Name()); ok && e.Type().IsRegular() {
				matches = append(matches, e.Name())
			}
		}
	}
	if len(matches) != 1 {
		return nil, nil
	}
	h, err := fileHashes(filepath.Join(serverDir, matches[0]))
	if err != nil {
		return nil, err
	}
	return &ServerFile{Path: matches[0], Hashes: h}, nil
}

// escapeGlob quotes the path.Match metacharacters in s.
func escapeGlob(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(`*?[\`, r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func fileHashes(file string) (modrinth.Hashes, error) {
	f, err := os.Open(file)
	if err != nil {
		return modrinth.Hashes{}, err
	}
	defer f.Close()
	h1 := sha1.New()
	h512 := sha512.New()
	_, err = io.Copy(io.MultiWriter(h1, h512), f)
	if err != nil {
		return modrinth.Hashes{}, err
	}
	return modrinth.Hashes{Sha1: hex.EncodeToString(h1.Sum(nil)), Sha512: hex.EncodeToString(h512.Sum(nil))}, nil
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
