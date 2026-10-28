package server

import (
	"reflect"
	"testing"
)

func TestNeoforgeVersionsFor(t *testing.T) {
	t.Parallel()
	all := []string{"21.0.167", "21.1.1", "21.1.265", "21.1.266-beta", "21.11.3", "21.11.48", "20.4.237"}
	tests := map[string][]string{
		"1.21.1":  {"21.1.1", "21.1.265"},
		"1.21":    {"21.0.167"},
		"1.21.11": {"21.11.3", "21.11.48"},
		"1.20.4":  {"20.4.237"},
		"1.19.2":  nil,
	}
	for mc, want := range tests {
		if got := neoforgeVersionsFor(all, mc); !reflect.DeepEqual(got, want) {
			t.Errorf("neoforgeVersionsFor(%s) = %v, want %v", mc, got, want)
		}
	}
}

// NeoForge for 1.20.1 is the fork of Forge, whose versions carry the
// Minecraft version as a prefix.
func TestNeoforgeVersionsFor1201(t *testing.T) {
	t.Parallel()
	all := []string{"1.20.1-47.1.7", "1.20.1-47.1.106", "1.20.1-47.1.107-beta"}

	got := neoforgeVersionsFor(all, "1.20.1")

	if want := []string{"47.1.7", "47.1.106"}; !reflect.DeepEqual(got, want) {
		t.Errorf("neoforgeVersionsFor(1.20.1) = %v, want %v", got, want)
	}
}

func TestNeoforgeSupports(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"1.20.1": true,
		"1.20.2": true,
		"1.21":   true,
		"1.21.1": true,
		"26.1":   true,
		"1.20":   false,
		"1.19.2": false,
		"1.16.5": false,
		"latest": true,
	}
	for mc, want := range tests {
		if got := neoforgeSupports(mc); got != want {
			t.Errorf("neoforgeSupports(%s) = %v, want %v", mc, got, want)
		}
	}
}
