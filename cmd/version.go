package cmd

import (
	"fmt"
	"log"
	"path"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version infos",
	Long:  `Extract and display the running binaries embedded version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		info, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Println(path.Base(info.Main.Path), info.Main.Version, info.Main.Sum)
		} else {
			log.Fatalln("Unable to extract build infos from running binary!")
		}
	},
}
