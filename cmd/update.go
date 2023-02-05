package cmd

import (
	"fmt"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
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
	// TODO flags: --start-server

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the deployed modpack",
	Long:  `Update the deployed modpacks config and mod files, creating backup files if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := GlobalOptions(cmd)

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

		err := os.MkdirAll(opts.ServerDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}

		archivePath := ""
		if util.PathIsFile(input) {
			archivePath = input

		} else if util.IsValidUrl(input) {
			fmt.Println("Downloading mrpack file from", args)
			file, err := requester.DefaultHttpClient.DownloadFile(input, opts.ServerDir, "")
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
			versions, err := modrinth.NewClient(opts.Host).GetProjectVersions(input, nil)
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
					file, err := requester.DefaultHttpClient.DownloadFile(files[i].Url, opts.ServerDir, "")
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

		index, err := mrpack.ReadIndex(archivePath)
		if err != nil {
			log.Fatalln(err)
		}

		newModPackInfo, err := update.GenerateModPackInfo(index)
		if err != nil {
			log.Fatalln(err)
		}

		for path := range newModPackInfo.Hashes {
			ok, err := util.PathIsSubpath(path, opts.ServerDir)
			if err != nil {
				log.Println(err.Error())
			}
			if err != nil || !ok {
				log.Fatalln("File path is not safe: " + path)
			}
		}

		err = newModPackInfo.Write(path.Join(opts.ServerDir, "modpack.json.update"))
		if err != nil {
			log.Fatalln(err)
		}
		oldModPackInfo, err := update.ReadModPackInfo(path.Join(opts.ServerDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}
		deleteFileInfo, updateFileInfo, err := update.CompareModPackInfo(*oldModPackInfo, *newModPackInfo)
		if err != nil {
			return
		}
		deletionActions := update.GetDeletionActions(deleteFileInfo, opts.ServerDir)
		updateActions := update.GetUpdateActions(updateFileInfo, opts.ServerDir)

		fmt.Printf("Would you like to update: [y/N]")
		var userInput string
		_, err = fmt.Scanln(&userInput)
		if err != nil {
			log.Fatalln(err)
		}
		if userInput != "y" {
			fmt.Println("Update process canceled.")
			return
		}

		err = update.ModPackDeleteDo(deletionActions, opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		var downloads []*requester.Download
		downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, opts.DownloadThreads, opts.RetryTimes)

		err = update.ModPackUpdateDo(updateActions, updateFileInfo.Hashes, opts.ServerDir, archivePath, downloadPools)
		if err != nil {
			log.Fatalln(err)
		}

		util.RemoveEmptyDirs(opts.ServerDir)

		err = os.Rename(path.Join(opts.ServerDir, "modpack.json.update"), path.Join(opts.ServerDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Done :) Have a nice day ✌️")
	},
}
