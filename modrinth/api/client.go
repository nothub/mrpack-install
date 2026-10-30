package api

import (
	"fmt"
	"github.com/nothub/mrpack-install/web"
	"log"
	"net/url"
	"strings"
)

type ModrinthClient struct {
	Http    *web.Client
	BaseUrl string
}

var Client *ModrinthClient

// NewClient accepts a bare host, which gets https, or a full URL.
func NewClient(host string) *ModrinthClient {
	base, err := BaseURL(host)
	if err != nil {
		log.Fatalln(err)
	}
	return &ModrinthClient{
		Http:    web.DefaultClient,
		BaseUrl: base,
	}
}

// BaseURL turns a labrinth host or URL into the API base URL, ending in a
// slash. A bare host gets https. It fails for anything requests can't be
// built on: no host, a scheme other than http or https, a query or a
// fragment.
func BaseURL(host string) (string, error) {
	h := strings.TrimSpace(host)
	if !strings.Contains(h, "://") {
		h = "https://" + h
	}
	u, err := url.Parse(h)
	if err != nil {
		return "", fmt.Errorf("invalid host %q: %w", host, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid host %q: scheme must be http or https", host)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid host %q: no host name", host)
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("invalid host %q: a query or fragment can't be a base URL", host)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/"
	return u.String(), nil
}
