package cmd

import (
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/spf13/cobra"
	"log"
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

/*
Behaviour:
    Deletions:
        PreDelete Three scenarios are possible:
        1.File does not exist notice
        2.File exists but hash value does not match, change the original file name to xxx.bak
        3.File exists and the hash value matches
    Updates:
        PreUpdate Three scenarios are possible:
        1.File does not exist
        2.File exists but hash value does not match, change the original file name to xxx.bak
        3.File exists and the hash value matches, remove the item from the queue
*/

var updateCmd = &cobra.Command{
	Use:   "update [<version>]",
	Short: "Update the deployed modpack",
	Long:  `Update the deployed modpacks files, creating backups if necessary.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := GetUpdateOpts(cmd)

		// user defined backup dir
		if opts.BackupDir != "" {
			backup.SetDir(opts.BackupDir)
		}

		version := ""
		if len(args) == 1 {
			version = args[0]
		}

		// TODO: get pack name from manifest

		index, zipPath := handleArgs(input, version, opts.ServerDir, opts.Host)

		update.Cmd(opts.ServerDir, opts.DlThreads, opts.DlRetries, index, zipPath)
	},
}
