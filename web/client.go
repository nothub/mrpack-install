package web

import (
	"fmt"
	"github.com/nothub/mrpack-install/internal/buildinfo"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	c  http.Client
	ua string
}

var DefaultClient = NewClient()

func NewClient() *Client {
	c := &Client{c: http.Client{}}
	c.c.Transport = NewTransport()
	c.ua = UserAgent()
	return c
}

func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   20 * time.Second,
		ResponseHeaderTimeout: 25 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}
}

func UserAgent() string {
	return fmt.Sprintf(
		"%s (+https://%s)",
		buildinfo.Name(),
		buildinfo.Module(),
	)
}

// SetProxy routes every request through the proxy at fixedURL. It only
// checks the URL: an unreachable proxy fails the first request.
func (c *Client) SetProxy(fixedURL string) error {
	proxy, err := url.Parse(fixedURL)
	if err != nil {
		return fmt.Errorf("invalid proxy %q: %w", fixedURL, err)
	}
	switch proxy.Scheme {
	case "http", "https", "socks5", "socks5h":
	default:
		return fmt.Errorf("invalid proxy %q: scheme must be http, https, socks5 or socks5h", fixedURL)
	}
	if proxy.Host == "" {
		return fmt.Errorf("invalid proxy %q: no host name", fixedURL)
	}

	transport := NewTransport()
	transport.Proxy = http.ProxyURL(proxy)
	c.c.Transport = transport
	return nil
}
