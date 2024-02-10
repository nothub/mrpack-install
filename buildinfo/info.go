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
var Release = "unknown"

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

func PrintInfos() {
	fmt.Printf("%s %s %s %s",
		Name(),
		Tag,
		rev,
		fmt.Sprintf("%s-%s-%s", Arch, Os, compiler),
	)
	if dirty {
		fmt.Printf(" %s", "DIRTY")
	}
	if Release != "true" {
		fmt.Printf(" %s", "PRE-RELEASE")
	}
	fmt.Print("\n")
}