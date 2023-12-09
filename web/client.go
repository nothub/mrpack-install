package web

import (
	"fmt"
	"hub.lol/mrpack-install/buildinfo"
	"net/http"
	"net/url"
	"strings"
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
		"%s/%s (+https://%s)",
		buildinfo.Name(),
		strings.TrimPrefix(buildinfo.Tag, "v"),
		buildinfo.Module(),
	)
}

func (c *Client) SetProxy(fixedURL string) error {
	proxy, err := url.Parse(fixedURL)
	if err != nil {
		return err
	}

	transport := NewTransport()
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
