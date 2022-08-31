package modrinth

import (
	"github.com/nothub/gorinth/http"
	"log"
	"net/url"
)

type ApiClient struct {
	Http    *http.Client
	BaseUrl string
}

func NewClient(host string) *ApiClient {
	client := ApiClient{
		Http: http.Instance,
	}
	u, err := url.Parse("https://" + host + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.BaseUrl = u.String()
	return &client
}
