package cmd

import (
	"log"

	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/spf13/cobra"
)

func init() {
	addSyncFlags(updateCmd)
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [<version>]",
	Short: "Update the installed modpack",
	Long: `Update the installed modpack to <version>, or to its latest release, the
same way mrpack-install <pack> <version> does, with the pack the server
dir's packstate.json records.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		version := ""
		if len(args) == 1 {
			version = args[0]
		}

		old, err := packstate.LoadPackState(serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		if old == nil {
			log.Fatalf("No installed pack in %q, install one with: mrpack-install <pack>\n", serverDir)
		}
		pack := old.ProjectId
		if pack == "" {
			pack = old.Slug
		}
		if pack == "" {
			log.Fatalf("The state in %q does not name a Modrinth project, update with: mrpack-install <pack>\n", serverDir)
		}

		syncPack(resolveProject(pack, version), old)
	},
}
