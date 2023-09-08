package cmd

import (
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/spf13/cobra"
	"log"
)

var (
	// local options
	backupDir string
)

func init() {
	// TODO flags: --start-server
	updateCmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory path")

	rootCmd.AddCommand(updateCmd)
}

/*
TODO: verify correct update behaviour
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
		// user defined backup dir
		if backupDir != "" {
			backup.SetDir(backupDir)
		}

		version := ""
		if len(args) == 1 {
			version = args[0]
		}

		state, err := packstate.LoadPackState(serverDir)
		if err != nil {
			log.Fatalln(err.Error())
		}

		index, zipPath := handleArgs(state.Slug, version, serverDir, host)

		update.Cmd(serverDir, dlThreads, dlRetries, index, zipPath, state)
	},
}
