package modrinth

import (
	"github.com/nothub/gorinth/http"
	"log"
	"net/url"
)

type ApiClient struct {
	Http *http.Client
}

func NewClient(host string) *ApiClient {
	client := ApiClient{
		Http: http.NewHttpClient(),
	}
	u, err := url.Parse("https://" + host + "/")
	if err != nil {
		log.Fatalln(err)
	}
	client.Http.BaseUrl = u.String()
	return &client
}
