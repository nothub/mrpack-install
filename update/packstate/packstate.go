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
	"path/filepath"
	"regexp"
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
}

// Options are the flags that decide which pack files a run installs.
type Options struct {
	OptionalSelect     []string `json:"optional-select"`
	OptionalDisableAll bool     `json:"optional-disable-all"`
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

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha512.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	archiveSha512 := hex.EncodeToString(h.Sum(nil))

	var state Schema
	state.SchemaVersion = SchemaVersion
	state.Name = index.Name
	state.Version = index.Version
	state.ArchiveSha512 = archiveSha512
	state.Options = opts
	state.Deps = index.Deps
	state.Hashes = make(map[string]modrinth.Hashes)

	state.Ids, err = ids(archiveSha512)
	if err != nil {
		log.Printf("Modrinth lookup failed, %s records no ids: %s\n", FileName, err)
		state.Ids = Ids{}
	}

	// download hashes
	for _, indexFile := range index.Files {
		if !opts.Selects(indexFile) {
			continue
		}
		state.Hashes[indexFile.Path] = indexFile.Hashes
	}

	// override hashes
	for p, hashes := range mrpack.OverrideHashes(path) {
		state.Hashes[p] = hashes
	}

	return &state, nil
}
