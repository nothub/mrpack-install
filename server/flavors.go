package server

import "strings"

type Flavor string

const (
	Vanilla Flavor = "vanilla"
	Fabric  Flavor = "fabric"
	Quilt   Flavor = "quilt"
	Forge   Flavor = "forge"
	Paper   Flavor = "paper"
)

func GetFlavor(flavor string) Flavor {
	switch strings.ToLower(flavor) {
	case "vanilla":
		return Vanilla
	case "fabric":
		return Fabric
	case "quilt":
		return Quilt
	case "forge":
		return Forge
	case "paper":
		return Paper
	}
	return ""
}

func (f Flavor) String() string {
	return string(f)
}
