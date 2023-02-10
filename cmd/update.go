package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
)

func init() {
	// TODO flags: --start-server

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the deployed modpack",
	Long:  `Update the deployed modpacks config and mod files, creating backup files if necessary.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		opts := GlobalOptions(cmd)

		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		index, archivePath := handleArgs(input, version, opts.ServerDir, opts.Host)

		newPackInfo, err := update.GenerateModPackInfo(index)
		if err != nil {
			log.Fatalln(err)
		}

		for path := range newPackInfo.Hashes {
			ok, err := util.PathIsSubpath(path, opts.ServerDir)
			if err != nil {
				log.Println(err.Error())
			}
			if err != nil || !ok {
				log.Fatalln("File path is not safe: " + path)
			}
		}

		fmt.Println("Updating:", index.Name)

		err = newPackInfo.Write(path.Join(opts.ServerDir, "modpack.json.update"))
		if err != nil {
			log.Fatalln(err)
		}
		oldPackInfo, err := update.ReadModPackInfo(path.Join(opts.ServerDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}
		deleteFileInfo, updateFileInfo, err := update.CompareModPackInfo(*oldPackInfo, *newPackInfo)
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

		err = update.ModPackUpdateDo(updateActions, updateFileInfo.Hashes, opts.ServerDir, archivePath, opts.DownloadThreads, opts.RetryTimes)
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
