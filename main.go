package main

import (
	"fmt"
	"github.com/nothub/gorinth/api"
	"log"
)

func main() {
	client := api.NewClient()
	info, err := client.LabrinthInfo()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(info.About)
	fmt.Println(info.Name, info.Version)
	fmt.Println(info.Documentation)
}
