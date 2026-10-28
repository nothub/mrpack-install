package api

import (
	"reflect"
	"testing"
)

func TestArrayAsParamLeavesArgumentAlone(t *testing.T) {
	t.Parallel()
	ids := []string{"a", "b"}

	got := arrayAsParam(ids)

	if got != `["a","b"]` {
		t.Errorf("arrayAsParam() = %s, want [\"a\",\"b\"]", got)
	}
	if !reflect.DeepEqual(ids, []string{"a", "b"}) {
		t.Errorf("argument = %v, want it unchanged", ids)
	}
}
