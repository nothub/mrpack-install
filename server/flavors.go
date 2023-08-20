package server

import (
	"github.com/samber/lo"
	"strings"
)

type Flavor string

const (
	Vanilla  Flavor = "vanilla"
	Fabric   Flavor = "fabric"
	Quilt    Flavor = "quilt"
	Forge    Flavor = "forge"
	NeoForge Flavor = "neoforge"
	Paper    Flavor = "paper"
	Unknown  Flavor = ""
)

var Flavors = []Flavor{
	Vanilla,
	Fabric,
	Quilt,
	Forge,
	NeoForge,
	Paper,
}

var FlavorNames = lo.Map(Flavors, func(f Flavor, _ int) string {
	return f.String()
})

func ToFlavor(s string) Flavor {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	for _, f := range Flavors {
		if s == f.String() {
			return f
		}
	}
	return Unknown
}

func (f Flavor) String() string {
	return string(f)
}
