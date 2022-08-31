package modrinth

import (
	"fmt"
	"strings"
)

func arrayAsParam(arr []string) string {
	for i := range arr {
		arr[i] = fmt.Sprintf("%q", arr[i])
	}
	return "[" + strings.Join(arr, ",") + "]"
}
