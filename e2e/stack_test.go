//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestStackServesGameVersions(t *testing.T) {
	res, err := http.Get(labrinthURL + "/v2/tag/game_version")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
}
