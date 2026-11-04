package cmd

import (
	"fmt"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	RootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping a Labrinth instance",
	Long:  `Connect to a Labrinth instance and display its statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Connecting to:", host)
		stats, err := modrinth.Client.Statistics()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("projects:", stats.Projects)
		fmt.Println("versions:", stats.Versions)
		fmt.Println("authors: ", stats.Authors)
		fmt.Println("files:   ", stats.Files)
	},
}
