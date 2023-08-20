package http

import (
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	c  http.Client
	ua string
}

var DefaultClient = newHTTPClient()

func newHTTPClient() *Client {
	c := &Client{c: http.Client{}}
	c.c.Transport = newTransport()
	c.ua = userAgent()
	return c
}

func userAgent() string {
	return fmt.Sprintf(
		"%s/%s (%s; %s) +https://%s",
		buildinfo.Name(),
		strings.TrimPrefix(buildinfo.Tag, "v"),
		buildinfo.Os,
		buildinfo.Arch,
		buildinfo.Module(),
	)
}

func newTransport() *http.Transport {
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

func (c *Client) SetProxy(fixedURL string) error {
	proxy, err := url.Parse(fixedURL)
	if err != nil {
		return err
	}

	transport := newTransport()
	transport.Proxy = http.ProxyURL(proxy)
	c.c.Transport = transport

	// Test proxy
	httpUrl := "https://api.modrinth.com/"
	response, err := c.c.Get(httpUrl)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
