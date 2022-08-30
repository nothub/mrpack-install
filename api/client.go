package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

type Client struct {
	UserAgent  string
	HTTPClient *http.Client
}

func NewClient() *Client {
	userAgent := "gorinth"
	info, ok := debug.ReadBuildInfo()
	if ok {
		userAgent = info.Main.Path + "/" + info.Main.Version
	}
	return &Client{
		UserAgent:  userAgent,
		HTTPClient: &http.Client{},
	}
}

func (client *Client) buildRequest(method string, url string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatalln(err)
	}

	request.Header.Set("User-Agent", client.UserAgent)
	request.Header.Set("Accept", "application/json")
	if request.Method == "POST" || request.Method == "PATCH" || request.Method == "PUT" {
		request.Header.Set("Content-Type", "application/json")
	}

	request.Close = true

	return request
}

func (client *Client) sendRequest(method string, url string, body io.Reader, result interface{}) error {
	response, err := client.HTTPClient.Do(client.buildRequest(method, url, body))
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
		var e Error
		if json.NewDecoder(response.Body).Decode(&e) != nil {
			return errors.New("http status " + strconv.Itoa(response.StatusCode))
		}
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + e.Error + " " + e.Description)
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + err.Error())
	}

	return nil
}
