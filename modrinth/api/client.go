package api

import (
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
	client := ModrinthClient{
		Http: web.DefaultClient,
	}
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}
	u, err := url.Parse(strings.TrimRight(host, "/") + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.BaseUrl = u.String()
	return &client
}
