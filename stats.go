//usr/bin/env -S go run "$0" "$@" ; exit
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

/*
  # https://stackoverflow.com/questions/4338358/github-can-i-see-the-number-of-downloads-for-a-repo/4339085#4339085
  # https://docs.github.com/en/rest/metrics/traffic
  # https://docs.github.com/en/rest/releases/releases
*/

type Release struct {
	Id          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Id            int       `json:"id"`
		Name          string    `json:"name"`
		State         string    `json:"state"`
		DownloadCount int       `json:"download_count"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	} `json:"assets"`
}

func main() {
	res, err := http.Get("https://api.github.com/repos/nothub/mrpack-install/releases")
	if err != nil {
		log.Fatalln(err.Error())
	}

	releases := make([]Release, 0)
	err = json.NewDecoder(res.Body).Decode(&releases)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("%++v", releases)
}
