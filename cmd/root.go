package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// TODO: flags
	// rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version infos")
	// rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "gorinth",
	Short: "Modrinth Modpack server deployment",
	Long:  `A cli application for installing Minecraft servers and Modrinth modpacks.`,
}

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
