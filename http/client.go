package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
)

// TODO: http client provider server -> 1 http client per target host

type ErrorModel interface {
	String() string
}

type Client struct {
	UserAgent  string
	BaseUrl    string
	HTTPClient *http.Client
}

func NewHttpClient() *Client {
	client := &Client{
		UserAgent:  "gorinth",
		HTTPClient: &http.Client{},
	}
	info, ok := debug.ReadBuildInfo()
	if ok {
		client.UserAgent = info.Main.Path + "/" + info.Main.Version
	}
	return client
}

func (client *Client) JsonRequest(method string, url string, body io.Reader, reponseModel interface{}, errorModel ErrorModel) error {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", client.UserAgent)
	request.Header.Set("Accept", "application/json")
	if request.Method == "POST" || request.Method == "PATCH" || request.Method == "PUT" {
		request.Header.Set("Content-Type", "application/json")
	}

	request.Close = true
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(response.Body)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		if errorModel == nil || json.NewDecoder(response.Body).Decode(&errorModel) != nil {
			return errors.New("http status " + strconv.Itoa(response.StatusCode))
		}
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + errorModel.String())
	}

	err = json.NewDecoder(response.Body).Decode(&reponseModel)
	if err != nil {
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + err.Error())
	}

	return nil
}
