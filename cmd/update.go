package cmd

import (
	"fmt"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"strings"
)

func init() {

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the deployed modpack",
	Long:  `Update the deployed modpacks config and mod files, creating backup files if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			err := cmd.Help()
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(1)
		}
		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}

		host, err := cmd.Flags().GetString("host")
		if err != nil {
			log.Fatalln(err)
		}
		serverDir, err := cmd.Flags().GetString("server-dir")
		if err != nil {
			log.Fatalln(err)
		}
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			log.Fatalln(err)
		}
		if proxy != "" {
			err := requester.DefaultHttpClient.SetProxy(proxy)
			if err != nil {
				log.Fatalln(err)
			}
		}
		downloadThreads, err := cmd.Flags().GetInt("download-threads")
		if err != nil || downloadThreads > 64 {
			downloadThreads = 8
			fmt.Println(err)
		}
		retryTimes, err := cmd.Flags().GetInt("retry-times")
		if err != nil {
			retryTimes = 3
			fmt.Println(err)
		}

		err = os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}

		archivePath := ""
		if util.PathIsFile(input) {
			archivePath = input

		} else if util.IsValidUrl(input) {
			fmt.Println("Downloading mrpack file from", args)
			file, err := requester.DefaultHttpClient.DownloadFile(input, serverDir, "")
			if err != nil {
				log.Fatalln(err)
			}
			archivePath = file
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					fmt.Println(err)
				}
			}(archivePath)
		} else { // input is project id or slug?
			versions, err := modrinth.NewClient(host).GetProjectVersions(input, nil)
			if err != nil {
				log.Fatalln(err)
			}

			// get files uploaded for specified version or latest stable if not specified
			var files []modrinth.File = nil
			for i := range versions {
				if version != "" {
					if versions[i].VersionNumber == version {
						files = versions[i].Files
						break
					}
				} else {
					if versions[i].VersionType == modrinth.ReleaseVersionType {
						files = versions[i].Files
						break
					}
				}
			}
			if len(files) == 0 {
				log.Fatalln("No files found for", input, version)
			}

			for i := range files {
				if strings.HasSuffix(files[i].Filename, ".mrpack") {
					fmt.Println("Downloading mrpack file from", files[i].Url)
					file, err := requester.DefaultHttpClient.DownloadFile(files[i].Url, serverDir, "")
					if err != nil {
						log.Fatalln(err)
					}
					archivePath = file
					break
				}
			}
			if archivePath == "" {
				log.Fatalln("No mrpack file found for", input, version)
			}
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					fmt.Println(err)
				}
			}(archivePath)
		}

		if archivePath == "" {
			log.Fatalln("An error occured!")
		}

		fmt.Println("Processing mrpack file", archivePath)

		var downloads []*requester.Download
		downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, downloadThreads, retryTimes)

		newModPackInfo, err := update.GenerateModPackInfo(archivePath)
		if err != nil {
			log.Fatalln(err)
		}

		for path, _ := range newModPackInfo.File {
			ok, err := util.PathIsSubpath(string(path), serverDir)
			if err != nil {
				log.Println(err.Error())
			}
			if err != nil || !ok {
				log.Fatalln("File path is not safe: " + path)
			}
		}

		err = newModPackInfo.Write(path.Join(serverDir, "modpack.json.update"))
		if err != nil {
			log.Fatalln(err)
		}
		oldModPackInfo, err := update.ReadModPackInfo(path.Join(serverDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}
		deleteFileInfo, updateFileInfo, err := update.CompareModPackInfo(*oldModPackInfo, *newModPackInfo)
		if err != nil {
			return
		}
		deleteList := update.PreDelete(deleteFileInfo, serverDir)
		updateList := update.PreUpdate(updateFileInfo, serverDir)

		fmt.Printf("Would you like to update: [y/N]")
		var userInput string
		_, err = fmt.Scanln(&userInput)
		if err != nil {
			log.Fatalln(err)
		}
		if userInput != "y" {
			return
		}

		err = update.ModPackDeleteDo(deleteList, serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		err = update.ModPackUpdateDo(updateList, updateFileInfo.File, serverDir, archivePath, downloadPools)
		if err != nil {
			log.Fatalln(err)
		}
		util.RemoveEmptyDir(serverDir)

		err = os.Rename(path.Join(serverDir, "modpack.json.update"), path.Join(serverDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Done :) Have a nice day ✌️")
	},
}
