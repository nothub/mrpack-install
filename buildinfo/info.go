package buildinfo

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"
	"strings"
)

const unknown string = "unknown"

var name = unknown
var module = unknown

var version = unknown
var revision = unknown
var dirty = false

var arch = unknown
var os = unknown
var compiler = unknown

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatalln("unable to read build info from binary")
		return
	}

	name = path.Base(bi.Main.Path)
	module = bi.Main.Path
	version = bi.Main.Version

	for _, kv := range bi.Settings {
		switch kv.Key {
		case "vcs.revision":
			revision = kv.Value[:8]
		case "vcs.modified":
			if kv.Value == "true" {
				dirty = true
			}
		case "GOARCH":
			arch = kv.Value
		case "GOOS":
			os = kv.Value
		case "-compiler":
			compiler = kv.Value
		}
	}

	if version == unknown {
		version = revision
	}
}

func Name() string {
	return name
}

func Module() string {
	return module
}

func Version() string {
	return version
}

func Revision() string {
	return revision
}

func PrintInfos() {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s %s", Name(), Version()))
	if Version() != Revision() {
		sb.WriteString(fmt.Sprintf(" %s", Revision()))
	}
	sb.WriteString(fmt.Sprintf(" %s-%s-%s", arch, os, compiler))
	if dirty {
		sb.WriteString(" DIRTY")
	}
	fmt.Println(sb.String())
}
