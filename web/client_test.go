package web

import "testing"

func TestSetProxy(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"http://127.0.0.1:3128", "https://proxy.example.org", "socks5://127.0.0.1:7890", "socks5h://user:pw@127.0.0.1:1080"} {
		if err := NewClient().SetProxy(p); err != nil {
			t.Errorf("SetProxy(%q) = %v, want nil", p, err)
		}
	}
	for _, p := range []string{"not a url", "127.0.0.1:3128", "ftp://127.0.0.1", "socks5://"} {
		if err := NewClient().SetProxy(p); err == nil {
			t.Errorf("SetProxy(%q) = nil, want an error", p)
		}
	}
}
