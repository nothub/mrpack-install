package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	modrinth "hub.lol/mrpack-install/modrinth/api"
	"log"
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping a Labrinth instance",
	Long:  `Connect to a Labrinth instance and display basic information.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Connecting to:", host)
		info, err := modrinth.Client.LabrinthInfo()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(info.About)
		fmt.Println(info.Name, info.Version)
		fmt.Println(info.Documentation)
	},
}
