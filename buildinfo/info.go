package buildinfo

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"
)

var name = "unknown"
var module = "unknown"

var Tag = "unknown"
var rev = "unknown"
var dirty = false

var Arch = "unknown"
var Os = "unknown"
var compiler = "unknown"

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
			Arch = kv.Value
		case "GOOS":
			Os = kv.Value
		case "-compiler":
			compiler = kv.Value
		}
	}
}

func Name() string {
	return name
}

func Module() string {
	return module
}

func SemVer() string {
	ver := fmt.Sprintf("%s-%s", Tag, rev)
	if dirty {
		ver = ver + "-dirty"
	}
	return ver
}

func PrintInfos() {
	fmt.Printf("%s %s %s\n",
		Name(),
		SemVer(),
		fmt.Sprintf("%s-%s-%s", Arch, Os, compiler),
	)
}
