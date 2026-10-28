package cmd

import (
	"errors"
	"log"
	"time"

	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/spf13/cobra"
)

var allowDowngrade bool

func init() {
	addSyncFlags(updateCmd)
	updateCmd.Flags().BoolVar(&allowDowngrade, "allow-downgrade", false, "Update to a version published before the installed one")
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [<version>]",
	Short: "Update the installed modpack",
	Long: `Update the installed modpack to <version>, or to its latest release, the
same way mrpack-install <pack> <version> does, with the pack the server
dir's packstate.json records. A version published before the installed
one needs --allow-downgrade.`,
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

		src := resolveProject(pack, version)
		if !allowDowngrade {
			refuseDowngrade(old, src.version)
		}
		syncPack(src, old)
	},
}

// refuseDowngrade exits when requested was published before the installed
// version. Pack version numbers follow no common scheme, so the publish
// dates decide. Without a recorded or still published installed version
// there is nothing to compare, and the update goes ahead.
func refuseDowngrade(old *packstate.Schema, requested *modrinth.Version) {
	if old.VersionId == "" || requested.Id == old.VersionId {
		return
	}
	installed, err := modrinth.Client.GetVersion(old.VersionId)
	if err != nil {
		log.Printf("Unable to look up the installed version %s, so it can't be compared: %s\n", old.VersionId, err)
		return
	}
	installedAt, err1 := time.Parse(time.RFC3339, installed.DatePublished)
	requestedAt, err2 := time.Parse(time.RFC3339, requested.DatePublished)
	if err := errors.Join(err1, err2); err != nil {
		log.Printf("Unable to compare publish dates: %s\n", err)
		return
	}
	if requestedAt.Before(installedAt) {
		log.Fatalf("%s was published before the installed %s, so this update would downgrade the server. Pass --allow-downgrade to do it anyway.\n",
			requested.VersionNumber, installed.VersionNumber)
	}
}
