package api

import (
	"github.com/nothub/mrpack-install/requester"
	"log"
	"net/url"
)

type ModrinthClient struct {
	Http    *requester.HTTPClient
	BaseUrl string
}

func NewClient(host string) *ModrinthClient {
	client := ModrinthClient{
		Http: requester.DefaultHttpClient,
	}
	u, err := url.Parse("https://" + host + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.BaseUrl = u.String()
	return &client
}
