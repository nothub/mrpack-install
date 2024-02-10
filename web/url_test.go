package web

import "testing"

func TestIsValidHttpUrl(t *testing.T) {
	type testCase struct {
		url  string
		want bool
	}

	cases := []testCase{
		{url: "C:\\Foo\\Bar", want: false},
		{url: "C://Foo//Bar", want: false},
		{url: "http://example.org", want: true},
		{url: "https://example.org", want: true},
		{url: "https://example.org\n", want: false},
		{url: "https://example.org/foo/bar", want: true},
		{url: "https://example.org/foo/bar?x=y", want: true},
	}

	for _, c := range cases {
		t.Run(c.url, func(t *testing.T) {
			if got := IsValidHttpUrl(c.url); got != c.want {
				t.Errorf("IsValidUrl() = %v, want %v", got, c.want)
			}
		})
	}
}
