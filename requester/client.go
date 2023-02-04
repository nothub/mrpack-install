package requester

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

var DefaultHttpClient = NewHTTPClient()

func (httpClient *HTTPClient) GetJson(url string, respModel interface{}, errModel error) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", httpClient.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Close = true

	res, err := httpClient.sendRequest(req)
	if err != nil {
		return err
	}

	defer func(Body io.Closer) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		if errModel == nil || json.NewDecoder(res.Body).Decode(&errModel) != nil {
			return errors.New("http status " + strconv.Itoa(res.StatusCode))
		}
		return errors.New("http status " + strconv.Itoa(res.StatusCode) + " - " + errModel.Error())
	}

	err = json.NewDecoder(res.Body).Decode(&respModel)
	if err != nil {
		return errors.New("http status " + strconv.Itoa(res.StatusCode) + " - " + err.Error())
	}

	return nil
}

func (httpClient *HTTPClient) DownloadFile(url string, downloadDir string, fileName string) (string, error) {
	// TODO: hashsum based local file cache

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", httpClient.UserAgent)
	request.Close = true

	response, err := httpClient.sendRequest(request)
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

func (httpClient *HTTPClient) sendRequest(request *http.Request) (*http.Response, error) {
	awaitRateLimits(request.Host)
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	updateRateLimits(response)
	return response, nil
}
