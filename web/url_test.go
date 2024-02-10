package web

import "testing"

func TestIsValidHttpUrl(t *testing.T) {
	type testCase struct {
		url  string
		want bool
	}

	cases := []testCase{
		{url: "http://example.org", want: true},
		{url: "http://example.org/path/to/page", want: true},
		{url: "http://example.org:8080", want: true},
		{url: "http://example.org:8080/path/to/page", want: true},

		{url: "https://example.org", want: true},
		{url: "https://example.org/foo/bar", want: true},
		{url: "https://example.org/foo/bar?x=y", want: true},
		{url: "https://example.org:8080", want: true},
		{url: "https://example.org:8080/path/to/page", want: true},
		{url: "https://example.org\n", want: false},

		{url: "example.org", want: false},

		{url: "ftp://example.org", want: false},

		{url: "relative/unix/path", want: false},
		{url: "relative/unix/path/", want: false},
		{url: "relative/unix/path/file.txt", want: false},

		{url: "relative\\windows\\path", want: false},
		{url: "relative\\windows\\path\\", want: false},
		{url: "relative\\windows\\path\\file.txt", want: false},

		{url: "/absolute/unix/path", want: false},
		{url: "/absolute/unix/path/", want: false},
		{url: "/absolute/unix/path/file.txt", want: false},

		{url: "C:\\absolute\\windows\\path", want: false},
		{url: "C:\\absolute\\windows\\path\\", want: false},
		{url: "C:\\absolute\\windows\\path\\file.txt", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			if got := IsValidHttpUrl(tc.url); got != tc.want {
				t.Errorf("IsValidUrl() = %v, want %v", got, tc.want)
			}
		})
	}
}
