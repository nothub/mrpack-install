package cmd

import (
	"github.com/nothub/mrpack-install/http"
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// TODO: flags
	// rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version infos")
	// rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "mrpack-install",
	Short: "Modrinth Modpack server deployment",
	Long:  `A cli application for installing Minecraft servers and Modrinth modpacks.`,

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

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
