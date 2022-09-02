package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
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
	if ok && info.Main.Path != "" {
		Instance.UserAgent = info.Main.Path + "/" + info.Main.Version
	}
}

func (client *Client) GetJson(url string, respModel interface{}, errModel ErrorModel) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
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

func (client *Client) DownloadFile(url string, downloadDir string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", client.UserAgent)
	request.Close = true

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}

	fileName := ""
	if response.Header.Get("content-disposition") != "" {
		matches := regexp.
			MustCompile("attachment; filename=\"(.*)\"").
			FindStringSubmatch(response.Header.Get("content-disposition"))
		if len(matches) > 1 {
			fileName = matches[1]
		}
	}
	if fileName == "" {
		fileName = path.Base(response.Request.URL.Path)
	}

	file, err := os.Create(path.Join(downloadDir, fileName))
	if err != nil {
		return "", err
	}
	err = file.Chmod(0644)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	err = response.Body.Close()
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}