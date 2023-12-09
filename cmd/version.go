package cmd

import (
	"github.com/spf13/cobra"
	"hub.lol/mrpack-install/buildinfo"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version infos",
	Long:  `Extract and display the running binaries embedded version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		buildinfo.PrintInfos()
	},
}
