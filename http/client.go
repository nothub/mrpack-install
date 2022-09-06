package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
)

type ErrorModel interface {
	String() string
}

var Instance = NewHTTPClient()

func (client *Client) GetJson(url string, respModel interface{}, errModel ErrorModel) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", client.UserAgent)
	request.Header.Set("Accept", "application/json")

	request.Close = true

	response, err := client.Do(request)
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

func (client *Client) DownloadFile(url string, downloadDir string, fileName string) (string, error) {
	// TODO: hashsum based local file cache

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", client.UserAgent)
	request.Close = true

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if fileName == "" && response.Header.Get("content-disposition") != "" {
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
	if fileName == "" {
		return "", errors.New("unable to determine file name")
	}

	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	joined := path.Join(downloadDir, fileName)
	file, err := os.Create(joined)
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
