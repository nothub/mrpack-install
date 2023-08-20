package web

import (
	"encoding/json"
	"encoding/xml"
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

func (c *Client) GetModel(url string, respModel interface{}, errModel error, decode func(body io.Reader, model interface{}) error) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	req.Close = true

	res, err := c.sendRequest(req)
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

	err = decode(res.Body, &respModel)
	if err != nil {
		return errors.New("http status " + strconv.Itoa(res.StatusCode) + " - " + err.Error())
	}

	return nil
}

func (c *Client) GetJson(url string, resModel interface{}, errModel error) error {
	return c.GetModel(url, resModel, errModel, func(i io.Reader, o interface{}) error {
		return json.NewDecoder(i).Decode(o)
	})
}

func (c *Client) GetXml(url string, resModel interface{}, errModel error) error {
	return c.GetModel(url, resModel, errModel, func(i io.Reader, o interface{}) error {
		return xml.NewDecoder(i).Decode(o)
	})
}

func (c *Client) DownloadFile(url string, downloadDir string, fileName string) (string, error) {
	// TODO: hashsum based local file cache

	// TODO: this needs to (silently?) overwrite existing files!

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", c.ua)
	request.Close = true

	response, err := c.sendRequest(request)
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

func (c *Client) sendRequest(request *http.Request) (*http.Response, error) {
	awaitRateLimits(request.Host)
	response, err := c.c.Do(request)
	if err != nil {
		return nil, err
	}
	updateRateLimits(response)
	return response, nil
}
