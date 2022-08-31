package cmd

import (
	"fmt"
	"log"

	"github.com/nothub/gorinth/modrinth"
	"github.com/spf13/cobra"
)

var host *string

func init() {
	host = pingCmd.Flags().String("host", "api.modrinth.com", "Host address")

	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping a Labrinth instance",
	Long:  `Connect to a Labrinth instance and display basic information.`,

	Run: func(cmd *cobra.Command, args []string) {
		client := modrinth.NewClient(*host)

		info, err := client.LabrinthInfo()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(info.About)
		fmt.Println(info.Name, info.Version)
		fmt.Println(info.Documentation)
	},
}
