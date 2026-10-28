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
