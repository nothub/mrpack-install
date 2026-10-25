package update

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/nothub/mrpack-install/files"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/update/packstate"
)

// Op is what a sync does to one path in the server dir.
type Op int

const (
	OpNone Op = iota
	OpWrite
	OpDelete
	OpBackupWrite
	OpBackupDelete
)

func (op Op) String() string {
	switch op {
	case OpNone:
		return "none"
	case OpWrite:
		return "write"
	case OpDelete:
		return "delete"
	case OpBackupWrite:
		return "backup+write"
	case OpBackupDelete:
		return "backup+delete"
	default:
		return fmt.Sprintf("Op(%d)", int(op))
	}
}

// Action is the planned Op for one pack-relative path. Recorded is set when
// the old state had a record for the path.
type Action struct {
	Path     string
	Op       Op
	Recorded bool
}

// Plan lists an Action for every path the old state records or the new
// pack ships, sorted by path, and whether the server installs.
type Plan struct {
	Actions       []Action
	InstallServer bool
}

// Compute plans the sync of serverDir from the old state, nil when there is
// none, to the pack with index and archive under opts. It reads file hashes
// from serverDir and changes nothing. An unsafe path in the old state or the
// new pack fails it before any action is planned.
func Compute(serverDir string, old *packstate.Schema, index *mrpack.Index, archive string, opts packstate.Options) (*Plan, error) {
	shipped := packstate.ShippedFiles(index, archive, opts)
	var recorded map[string]modrinth.Hashes
	if old != nil {
		recorded = old.Hashes
	}

	var paths []string
	for p := range shipped {
		paths = append(paths, p)
	}
	for p := range recorded {
		if _, ok := shipped[p]; !ok && !opts.Excludes(p) {
			paths = append(paths, p)
		}
	}
	slices.Sort(paths)

	full := make(map[string]string, len(paths))
	for _, p := range paths {
		if p == stagingDir || strings.HasPrefix(p, stagingDir+"/") {
			return nil, fmt.Errorf("file path is reserved for staging downloads: %q", p)
		}
		joined, err := files.InDir(serverDir, p)
		if err != nil {
			return nil, err
		}
		full[p] = joined
	}

	install, err := installServer(serverDir, old, index.Deps, opts.ServerFile)
	if err != nil {
		return nil, err
	}
	plan := Plan{InstallServer: install || opts.ForceSync}
	for _, p := range paths {
		disk, err := hashFile(full[p])
		if err != nil {
			return nil, err
		}
		oldHashes, isRecorded := recorded[p]
		newHashes, isShipped := shipped[p]
		// A pack that renames mods/Foo.jar to mods/foo.jar finds the old
		// file under the new name on a case-insensitive filesystem. Its
		// record still tells whether it was modified.
		comparable := isRecorded
		if isShipped && !isRecorded {
			for q, h := range recorded {
				if q != p && strings.EqualFold(q, p) && sameFile(full[q], full[p]) {
					oldHashes, comparable = h, true
					break
				}
			}
		}

		action := Action{Path: p, Recorded: isRecorded}
		switch {
		case isShipped && disk != nil && matches(*disk, newHashes):
			action.Op = OpNone
		case isShipped && (disk == nil || comparable && matches(*disk, oldHashes)):
			action.Op = OpWrite
		case isShipped:
			action.Op = OpBackupWrite
		case disk == nil:
			action.Op = OpNone
		case matches(*disk, oldHashes):
			action.Op = OpDelete
		default:
			action.Op = OpBackupDelete
		}
		plan.Actions = append(plan.Actions, action)
	}

	plan.Actions = slices.DeleteFunc(plan.Actions, func(a Action) bool {
		if a.Op != OpDelete && a.Op != OpBackupDelete {
			return false
		}
		for p := range shipped {
			if p != a.Path && strings.EqualFold(p, a.Path) && sameFile(full[p], full[a.Path]) {
				return true
			}
		}
		return false
	})

	return &plan, nil
}

// Describe writes one line per action that changes the server dir, in
// path order, and whether the server installs.
func (plan *Plan) Describe(w io.Writer) error {
	for _, a := range plan.Actions {
		if a.Op == OpNone {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s %s\n", a.Op, a.Path); err != nil {
			return err
		}
	}
	install := "no"
	if plan.InstallServer {
		install = "yes"
	}
	_, err := fmt.Fprintf(w, "install server: %s\n", install)
	return err
}

// NoOp reports whether syncing the archive with SHA-512 sum under opts
// would change nothing: the old state records the same archive and the same
// options, the recorded server jar is in place, and ForceSync is not set.
func NoOp(serverDir string, old *packstate.Schema, sum string, opts packstate.Options) (bool, error) {
	if opts.ForceSync || old == nil || sum == "" || old.ArchiveSha512 != sum || !old.Options.Equal(opts) {
		return false, nil
	}
	// The same archive has the same dependencies, so this only checks the
	// server jar.
	install, err := installServer(serverDir, old, old.Deps, opts.ServerFile)
	return !install, err
}

// installServer reports whether the server needs installing: there is no
// old state, the Minecraft or a loader version changed, the recorded server
// jar is missing, or --server-file names another file than the recorded one.
func installServer(serverDir string, old *packstate.Schema, deps mrpack.Deps, serverFile string) (bool, error) {
	if old == nil || old.Deps != deps {
		return true, nil
	}
	recorded := old.ServerFile
	if serverFile != "" && (recorded == nil || recorded.Path != serverFile) {
		return true, nil
	}
	if recorded == nil || recorded.Path == "" {
		return false, nil
	}
	full, err := files.InDir(serverDir, recorded.Path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(full)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	return false, err
}

// writes returns the paths the plan writes.
func (plan *Plan) writes() map[string]bool {
	writes := make(map[string]bool)
	for _, a := range plan.Actions {
		if a.Op == OpWrite || a.Op == OpBackupWrite {
			writes[a.Path] = true
		}
	}
	return writes
}

// matches compares disk against h by the strongest hash h has.
func matches(disk modrinth.Hashes, h modrinth.Hashes) bool {
	if h.Sha512 != "" {
		return strings.EqualFold(disk.Sha512, h.Sha512)
	}
	if h.Sha1 != "" {
		return strings.EqualFold(disk.Sha1, h.Sha1)
	}
	return false
}

// hashFile returns the hashes of the file at path, or nil when there is no
// file.
func hashFile(path string) (*modrinth.Hashes, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", path)
	}
	h1 := sha1.New()
	h512 := sha512.New()
	if _, err := io.Copy(io.MultiWriter(h1, h512), f); err != nil {
		return nil, fmt.Errorf("hashing %s: %w", path, err)
	}
	return &modrinth.Hashes{Sha1: hex.EncodeToString(h1.Sum(nil)), Sha512: hex.EncodeToString(h512.Sum(nil))}, nil
}

// sameFile reports whether paths a and b name one file, as two names that
// differ only in case do on a case-insensitive filesystem.
func sameFile(a string, b string) bool {
	ia, err := os.Stat(a)
	if err != nil {
		return false
	}
	ib, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ia, ib)
}
