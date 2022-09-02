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

type ErrorModel interface {
	String() string
}

type Client struct {
	UserAgent  string
	HTTPClient *http.Client
}

var Instance *Client = nil

func init() {
	Instance = &Client{
		UserAgent:  "gorinth",
		HTTPClient: &http.Client{},
	}
	info, ok := debug.ReadBuildInfo()
	if ok {
		Instance.UserAgent = info.Main.Path + "/" + info.Main.Version
	}
}

func (client *Client) GetJson(url string, body io.Reader, respModel interface{}, errModel ErrorModel) error {
	request, err := http.NewRequest("GET", url, body)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", client.UserAgent)
	request.Header.Set("Accept", "application/json")

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
		if errModel == nil || json.NewDecoder(response.Body).Decode(&errModel) != nil {
			return errors.New("http status " + strconv.Itoa(response.StatusCode))
		}
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + errModel.String())
	}

	err = json.NewDecoder(response.Body).Decode(&respModel)
	if err != nil {
		return errors.New("http status " + strconv.Itoa(response.StatusCode) + " - " + err.Error())
	}

	return nil
}
