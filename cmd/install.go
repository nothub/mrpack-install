package cmd

import (
	"github.com/nothub/gorinth/http"
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"os"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a Modrinth Modpack",
	Long:  `TODO`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			// TODO: cobra arg checks
			log.Fatalln("argghh")
		}

		if _, err := os.Stat(args[0]); err == nil {
			// arg is existing file path

		} else if _, err := url.Parse(args[0]); err != nil {
			// arg is valid url

			file, err := http.Instance.DownloadFile(args[0], ".")
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Stored mrpack to:", file)

		} else {
			// arg is project id?
			// TODO
		}

		// TODO
	},
}
