//usr/bin/env -S go run "$0" "$@" ; exit
//go:build exclude

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Release struct {
	Id          int       `json:"id"`
	Tag         string    `json:"tag_name"`
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

	var releases []Release

	err = json.NewDecoder(res.Body).Decode(&releases)
	if err != nil {
		log.Fatalln(err.Error())
	}

	total := 0
	for i, _ := range releases {
		release := releases[len(releases)-1-i]
		fmt.Printf("%v:\n", release.Tag)
		combined := 0
		for _, asset := range release.Assets {
			trimmed := strings.TrimPrefix(asset.Name, "mrpack-install-")
			trimmed = strings.TrimSuffix(trimmed, ".exe")
			fmt.Printf("  - %v: %v\n", trimmed, asset.DownloadCount)
			combined = combined + asset.DownloadCount
		}
		fmt.Printf("  - combined: %v\n\n", combined)
		total = total + combined
	}
	fmt.Printf("total: %v\n", total)
}
