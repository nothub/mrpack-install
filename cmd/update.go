package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
)

func init() {
	// TODO flags: --start-server
	updateCmd.Flags().String("backup-dir", "", "Backup directory path")

	rootCmd.AddCommand(updateCmd)
}

type UpdateOpts struct {
	*GlobalOpts
	BackupDir string
}

func GetUpdateOpts(cmd *cobra.Command) *UpdateOpts {
	var opts UpdateOpts
	opts.GlobalOpts = GlobalOptions(cmd)

	backupDir, err := cmd.Flags().GetString("backup-dir")
	if err != nil {
		log.Fatalln(err)
	}
	opts.BackupDir = backupDir

	return &opts
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the deployed modpack",
	Long:  `Update the deployed modpacks files, creating backups if necessary.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		opts := GetUpdateOpts(cmd)

		// user defined backup dir
		if opts.BackupDir != "" {
			backup.SetDir(opts.BackupDir)
		}

		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		index, zipPath := handleArgs(input, version, opts.ServerDir, opts.Host)
		fmt.Println("Updating:", index.Name)

		newPackInfo, err := update.BuildPackState(zipPath)
		if err != nil {
			log.Fatalln(err)
		}

		for filePath := range newPackInfo.Hashes {
			util.AssertPathSafe(filePath, opts.ServerDir)
		}

		err = newPackInfo.Save(path.Join(opts.ServerDir, "modpack.json.update"))
		if err != nil {
			log.Fatalln(err)
		}
		oldPackInfo, err := update.LoadPackState(path.Join(opts.ServerDir, "modpack.json"))
		if err != nil {
			log.Fatalln(err)
		}

		// TODO: clean this up (phase 1: collect all required actions  phase 2: execute backups) (ignore all deletions here and just overwrite later on?)
		deletions, updates, err := update.CompareModPackInfo(*oldPackInfo, *newPackInfo)
		if err != nil {
			return
		}
		deletionActions := update.GetDeletionActions(deletions, opts.ServerDir)
		updateActions := update.GetUpdateActions(updates, opts.ServerDir)

		reportChanges(deletionActions, updateActions)
		if !askContinue() {
			fmt.Println("Update process canceled.")
			return
		}

		err = update.HandleOldFiles(deletionActions, opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		err = update.Do(updateActions, updates.Hashes, opts.ServerDir, zipPath, opts.DownloadThreads, opts.RetryTimes)
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

func reportChanges(deletions update.Actions, updates update.Actions) {
	var changes update.Actions
	for filePath, strategy := range deletions {
		changes[filePath] = strategy
	}
	for filePath, strategy := range updates {
		changes[filePath] = strategy
	}
	// TODO: include overrides in change report

	fmt.Printf("The following %v changes will be applied:\n", len(changes))
	for filePath, strategy := range changes {
		switch strategy {
		case update.Delete:
			log.Printf("Delete and replace: %s\n", filePath)
		case update.Backup:
			log.Printf("Backup and replace: %s\n", filePath)
		case update.NoOp:
			log.Printf("Create new file:    %s\n", filePath)
		}
	}
}

func askContinue() bool {
	fmt.Printf("Would you like to continue? [y/n]")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatalln(err)
	}
	if input == "y" {
		return true
	}
	fmt.Println("Stopping process.")
	return false
}
