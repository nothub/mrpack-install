package http

import (
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	"net/http"
	"net/url"
	"runtime/debug"
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

	c.ua = fmt.Sprintf("%s/%s", "mrpack-install", buildinfo.Version)
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Path != "" {
		c.ua = fmt.Sprintf("%s (+https://%s)", c.ua, info.Main.Path)
	}

	return c
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
