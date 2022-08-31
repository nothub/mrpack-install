package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Prepare a server environment",
	Long:  `TODO`,

	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"vanilla", "fabric", "forge", "quilt", "paper", "spigot"},

	Run: func(cmd *cobra.Command, args []string) {
		flavour := args[0]
		log.Println("Installing", flavour, "server...")
	},
}
