package cmd

import (
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	// local options
	minecraftVersion string
	flavorVersion    string
)

func init() {
	serverCmd.Flags().StringVar(&minecraftVersion, "minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().StringVar(&flavorVersion, "flavor-version", "latest", "Flavor version")

	// TODO: --eula
	// TODO: --op <uuid>...
	// TODO: --whitelist <uuid>...
	// TODO: --start-server

	rootCmd.AddCommand(serverCmd)

	cobra.OnInitialize(func() {
		// --minecraft-version
		if minecraftVersion == "" || minecraftVersion == "latest" {
			latestMinecraftVersion, err := mojang.LatestRelease()
			if err != nil {
				log.Fatalln(err)
			}
			minecraftVersion = latestMinecraftVersion
		}
		minecraftVersion = minecraftVersion
	})
}

var serverCmd = &cobra.Command{
	Use:   "server ( " + strings.Join(server.FlavorNames(), " | ") + " )",
	Short: "Prepare a plain server environment",
	Long:  `Download and configure one of several Minecraft server flavors.`,
	Example: `  mrpack-install server fabric --server-dir fabric-srv
  mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: server.FlavorNames(),
	Run: func(cmd *cobra.Command, args []string) {
		err := os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Chdir(serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		flavor := server.ToFlavor(args[0])
		inst, err := server.NewInstaller(flavor, minecraftVersion, flavorVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = inst.Install(serverDir, serverFile)
		if err != nil {
			log.Fatalln(err)
		}
	},
}
