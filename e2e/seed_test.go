//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nothub/mrpack-install/modrinth/api"
)

func TestSeedServesLatestGameVersion(t *testing.T) {
	got, err := api.NewClient(labrinthURL).GetLatestGameVersion()
	if err != nil {
		t.Fatal(err)
	}
	// The newest release in the vendored seed dump.
	if want := "1.21.8"; got != want {
		t.Errorf("GetLatestGameVersion() = %q, want %q", got, want)
	}
}

func TestSeedPATsAuthenticate(t *testing.T) {
	tests := []struct {
		pat  string
		role string
	}{
		{userPAT, "developer"},
		{adminPAT, "admin"},
	}
	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, labrinthURL+"/v2/user", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", tt.pat)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
			}
			var user struct {
				Role string `json:"role"`
			}
			if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
				t.Fatal(err)
			}
			if user.Role != tt.role {
				t.Errorf("role = %q, want %q", user.Role, tt.role)
			}
		})
	}
}
