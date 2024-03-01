package buildinfo

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"
)

const unknown string = "unknown"

var name = unknown
var module = unknown

var version = unknown
var rev = unknown
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

	for _, kv := range bi.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value[:8]
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
		version = rev
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

func PrintInfos() {
	fmt.Printf("%s %s %s",
		Name(),
		Version(),
		fmt.Sprintf("%s-%s-%s", arch, os, compiler),
	)
	if dirty {
		fmt.Printf(" %s", "DIRTY")
	}
	fmt.Print("\n")
}
