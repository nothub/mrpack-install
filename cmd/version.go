package cmd

import (
	"github.com/nothub/mrpack-install/buildinfo"
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
		buildinfo.PrintInfos()
	},
}
