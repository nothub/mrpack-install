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
		opts := GlobalOptions(cmd)

		log.Println("Connecting to:", opts.Host)
		client := modrinth.NewClient(opts.Host)
		info, err := client.LabrinthInfo()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(info.About)
		fmt.Println(info.Name, info.Version)
		fmt.Println(info.Documentation)
	},
}
