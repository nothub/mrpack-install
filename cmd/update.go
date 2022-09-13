package cmd

import (
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
