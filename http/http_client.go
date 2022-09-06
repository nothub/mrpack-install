package http

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime/debug"
	"time"
)

type Client struct {
	http.Client
	insecureSkipVerify bool
	UserAgent          string
	transport          *http.Transport
}

// TODO: global lookup map host -> ratelimit hits left and sleep wait strategy

func NewHTTPClient() *Client {
	client := &Client{
		Client:    http.Client{},
		UserAgent: "mrpack-install",
	}
	client.Client.Jar, _ = cookiejar.New(nil)
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Path != "" {
		client.UserAgent = info.Main.Path + "/" + info.Main.Version
	}
	return client
}

func (client *Client) lazyInit() {
	if client.transport == nil {
		client.transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			TLSHandshakeTimeout:   20 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false, // gzip
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 25 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		}
		client.Client.Transport = client.transport
	}
}

func (client *Client) SetUserAgent(ua string) {
	client.UserAgent = ua
}

func (client *Client) SetCookiejar(jar http.CookieJar) {
	client.Client.Jar = jar
}

func (client *Client) ResetCookiejar() {
	client.Jar, _ = cookiejar.New(nil)
}

func (client *Client) SetProxy(CustomProxy string) error {
	client.lazyInit()
	proxy, err := url.Parse(CustomProxy)
	if err != nil {
		return err
	}

	client.transport.Proxy = http.ProxyURL(proxy)

	// Test proxy
	httpUrl := "https://api.modrinth.com/"
	response, err := client.Get(httpUrl)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return err
	}
	return nil
}

func (client *Client) SetInsecureSkipVerify(b bool) {
	client.lazyInit()
	client.transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: b,
	}
}

func (client *Client) SetKeepAlive(b bool) {
	client.lazyInit()
	client.transport.DisableKeepAlives = !b
}

func (client *Client) SetGzip(b bool) {
	client.lazyInit()
	client.transport.DisableCompression = !b
}

func (client *Client) SetResponseHeaderTimeout(t time.Duration) {
	client.lazyInit()
	client.transport.ResponseHeaderTimeout = t
}

func (client *Client) SetTLSHandshakeTimeout(t time.Duration) {
	client.lazyInit()
	client.transport.TLSHandshakeTimeout = t
}

func (client *Client) SetTimeout(t time.Duration) {
	client.Client.Timeout = t
}
