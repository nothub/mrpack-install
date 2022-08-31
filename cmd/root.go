package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// global flags
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
	rootCmd.PersistentFlags().Bool("dumm", true, "Use dumm for dumm")

	// local flags
	rootCmd.Flags().BoolP("toggle", "t", false, "toggle message for toggle flag")
}

var rootCmd = &cobra.Command{
	Use:   "gorinth",
	Short: "Modrinth Modpack server deployment",
	Long: `A longer description that spans multiple lines and likely
contains examples and usage of using your application.
For example:
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
