package requester

import (
	"crypto/tls"
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime/debug"
	"time"
)

type HTTPClient struct {
	http.Client
	UserAgent string
	transport *http.Transport
}

func NewHTTPClient() *HTTPClient {
	httpClient := &HTTPClient{
		Client: http.Client{},
	}

	httpClient.Client.Jar, _ = cookiejar.New(nil)

	httpClient.UserAgent = fmt.Sprintf("%s/%s", "mrpack-install", buildinfo.Version)
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Path != "" {
		httpClient.UserAgent = fmt.Sprintf("%s (+https://%s)", httpClient.UserAgent, info.Main.Path)
	}

	return httpClient
}

func (httpClient *HTTPClient) lazyInit() {
	if httpClient.transport == nil {
		httpClient.transport = &http.Transport{
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
		httpClient.Client.Transport = httpClient.transport
	}
}

func (httpClient *HTTPClient) SetUserAgent(ua string) {
	httpClient.UserAgent = ua
}

func (httpClient *HTTPClient) SetCookiejar(jar http.CookieJar) {
	httpClient.Client.Jar = jar
}

func (httpClient *HTTPClient) ResetCookiejar() {
	httpClient.Jar, _ = cookiejar.New(nil)
}

func (httpClient *HTTPClient) SetProxy(CustomProxy string) error {
	httpClient.lazyInit()
	proxy, err := url.Parse(CustomProxy)
	if err != nil {
		return err
	}

	httpClient.transport.Proxy = http.ProxyURL(proxy)

	// Test proxy
	httpUrl := "https://api.modrinth.com/"
	response, err := httpClient.Get(httpUrl)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return err
	}
	return nil
}

func (httpClient *HTTPClient) SetInsecureSkipVerify(b bool) {
	httpClient.lazyInit()
	httpClient.transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: b,
	}
}

func (httpClient *HTTPClient) SetKeepAlive(b bool) {
	httpClient.lazyInit()
	httpClient.transport.DisableKeepAlives = !b
}

func (httpClient *HTTPClient) SetGzip(b bool) {
	httpClient.lazyInit()
	httpClient.transport.DisableCompression = !b
}

func (httpClient *HTTPClient) SetResponseHeaderTimeout(t time.Duration) {
	httpClient.lazyInit()
	httpClient.transport.ResponseHeaderTimeout = t
}

func (httpClient *HTTPClient) SetTLSHandshakeTimeout(t time.Duration) {
	httpClient.lazyInit()
	httpClient.transport.TLSHandshakeTimeout = t
}

func (httpClient *HTTPClient) SetTimeout(t time.Duration) {
	httpClient.Client.Timeout = t
}
