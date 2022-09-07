package server

import (
	"errors"
	"github.com/nothub/mrpack-install/requester"
	"strconv"
)

type Paper struct {
	MinecraftVersion string
	PaperVersion     string
}

func (supplier *Paper) GetUrl() (string, error) {
	var response struct {
		Builds []struct {
			Id        int    `json:"build"`
			Channel   string `json:"channel"`
			Downloads struct {
				Application struct {
					Name   string `json:"name"`
					Sha256 string `json:"sha256"`
				} `json:"application"`
			} `json:"downloads"`
		} `json:"builds"`
	}
	err := requester.DefaultHttpClient.GetJson("https://api.papermc.io/v2/projects/paper/versions/"+supplier.MinecraftVersion+"/builds", &response, nil)
	if err != nil {
		return "", err
	}
	for i := range response.Builds {
		i = len(response.Builds) - 1 - i
		if response.Builds[i].Channel == "default" {
			return "https://api.papermc.io/v2/projects/paper/versions/" + supplier.MinecraftVersion + "/builds/" + strconv.Itoa(response.Builds[i].Id) + "/downloads/" + response.Builds[i].Downloads.Application.Name, nil
		}
	}
	return "", errors.New("no stable paper release found")
}
