package api

import (
	"fmt"
	"strings"
)

func arrayAsParam(arr []string) string {
	quoted := make([]string, len(arr))
	for i, s := range arr {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ",") + "]"
}
