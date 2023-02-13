package api

import (
	"github.com/nothub/mrpack-install/http"
	"log"
	"net/url"
)

type ModrinthClient struct {
	Http    *http.Client
	BaseUrl string
}

func NewClient(host string) *ModrinthClient {
	client := ModrinthClient{
		Http: http.DefaultClient,
	}
	u, err := url.Parse("https://" + host + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.BaseUrl = u.String()
	return &client
}
