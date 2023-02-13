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

		update.Cmd(opts.ServerDir, opts.DlThreads, opts.DlRetries, index, zipPath)
	},
}
