package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a Modrinth Modpack",
	Long:  `TODO`,

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("TODO")
	},
}
