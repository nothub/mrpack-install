package api

import "testing"

func TestNewClient(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"api.modrinth.com", "https://api.modrinth.com/"},
		{"api.modrinth.com/", "https://api.modrinth.com/"},
		{"http://labrinth.lan:8000", "http://labrinth.lan:8000/"},
		{"https://labrinth.example.org/", "https://labrinth.example.org/"},
		{"https://labrinth.example.org///", "https://labrinth.example.org/"},
		{"https://example.org/labrinth", "https://example.org/labrinth/"},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := NewClient(tt.host).BaseUrl; got != tt.want {
				t.Errorf("NewClient(%q).BaseUrl = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}
