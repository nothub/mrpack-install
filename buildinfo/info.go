package buildinfo

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"
)

const unknown string = "unknown"

var (
	name   = unknown
	module = unknown

	version = unknown
	commit  = unknown
	date    = unknown
	tool    = unknown

	arch     = unknown
	os       = unknown
	compiler = unknown
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatalln("unable to read build info from binary")
		return
	}

	name = path.Base(bi.Main.Path)
	module = bi.Main.Path

	var dirty = false

	for _, kv := range bi.Settings {
		switch kv.Key {
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

	if dirty {
		version = version + "+DIRTY"
	}
}

func Name() string {
	return name
}

func Module() string {
	return module
}

func Print() {
	fmt.Printf("version:    %s\n", version)
	fmt.Printf("target:     %s-%s-%s\n", arch, os, compiler)
	fmt.Printf("built at:   %s\n", date)
	fmt.Printf("built from: %s\n", commit)
	fmt.Printf("built with: %s\n", tool)
}
