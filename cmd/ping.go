package cmd

import (
	"fmt"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/spf13/cobra"
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
		host, err := rootCmd.Flags().GetString("host")
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Connecting to:", host)
		client := modrinth.NewClient(host)
		info, err := client.LabrinthInfo()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(info.About)
		fmt.Println(info.Name, info.Version)
		fmt.Println(info.Documentation)
	},
}
