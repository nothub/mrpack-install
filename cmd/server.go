package cmd

import (
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().String("flavor-version", "latest", "Flavor version")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server (vanilla | fabric | quilt | forge | paper | spigot)",
	Short: "Prepare a plain server environment",
	Long:  `Download and configure one of several Minecraft server flavors.`,
	Example: `  mrpack-install server fabric --server-dir fabric-srv
  mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar`,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"vanilla", "fabric", "quilt", "forge", "paper", "spigot"},
	Run: func(cmd *cobra.Command, args []string) {
		minecraftVersion, err := cmd.Flags().GetString("minecraft-version")
		if err != nil {
			log.Fatalln(err)
		}
		flavorVersion, err := cmd.Flags().GetString("flavor-version")
		if err != nil {
			log.Fatalln(err)
		}
		serverDir, err := cmd.Flags().GetString("server-dir")
		if err != nil {
			log.Fatalln(err)
		}
		serverFile, err := cmd.Flags().GetString("server-file")
		if err != nil {
			log.Fatalln(err)
		}
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			log.Fatalln(err)
		}
		if proxy != "" {
			err := requester.DefaultHttpClient.SetProxy(proxy)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if minecraftVersion == "" || minecraftVersion == "latest" {
			latestMinecraftVersion, err := mojang.LatestRelease()
			if err != nil {
				log.Fatalln(err)
			}
			minecraftVersion = latestMinecraftVersion
		}

		err = os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}

		flavor := args[0]
		provider, err := server.NewProvider(flavor, minecraftVersion, flavorVersion)
		if err != nil {
			log.Fatalln(err)
		}
		err = provider.Provide(serverDir, serverFile)
		if err != nil {
			log.Fatalln(err)
		}
	},
}
