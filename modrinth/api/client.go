package api

import (
	"github.com/nothub/mrpack-install/web"
	"log"
	"net/url"
)

type ModrinthClient struct {
	Http    *web.Client
	BaseUrl string
}

func NewClient(host string) *ModrinthClient {
	client := ModrinthClient{
		Http: web.DefaultClient,
	}
	u, err := url.Parse("https://" + host + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.BaseUrl = u.String()
	return &client
}
