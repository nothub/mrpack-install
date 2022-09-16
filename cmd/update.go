package cmd

import (
	"github.com/nothub/mrpack-install/update/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the server",
	Long:  `Use file's hash and compare,Update the config and mods file'`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

// PreDelete Three scenarios
// 1.File does not exist Notice
// 2.File exists but hash value does not match,Change the original file name to xxx.bak
// 3.File exists and the hash value matches
func PreDelete(deleteList *model.ModPackInfo) error {
	return nil
}

// PreUpdate Three scenarios
// 1.File does not exist
// 2.File exists but hash value does not match,Change the original file name to xxx.bak
// 3.File exists and the hash value matches,Remove the item from the queue
func PreUpdate(updateList *model.ModPackInfo) error {
	return nil
}
