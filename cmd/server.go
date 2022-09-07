package cmd

import (
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().String("flavor-version", "latest", "Flavor build version")
	serverCmd.Flags().String("server-dir", "mc", "Server directory path")
	serverCmd.Flags().String("server-file", "", "Server jar file name")
	/*
	   TODO: eula flag
	   TODO: ops flag
	   TODO: whitelist flags
	*/
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:       "server (vanilla | fabric | quilt | forge | paper | spigot)",
	Short:     "Prepare a server environment",
	Long:      `Download and configure one of several Minecraft server flavors.`,
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

		if minecraftVersion == "" || minecraftVersion == "latest" {
			latestMinecraftVersion, err := mojang.LatestVersion()
			if err != nil {
				log.Fatalln(err)
			}
			minecraftVersion = latestMinecraftVersion
		}

		var provider server.Provider = nil
		switch args[0] {
		case "vanilla":
			provider = &server.Vanilla{
				MinecraftVersion: minecraftVersion,
			}
		case "fabric":
			provider = &server.Fabric{
				MinecraftVersion: minecraftVersion,
				FabricVersion:    flavorVersion,
			}
		case "quilt":
			provider = &server.Quilt{
				MinecraftVersion: minecraftVersion,
				QuiltVersion:     flavorVersion,
			}
		case "forge":
			provider = &server.Forge{
				MinecraftVersion: minecraftVersion,
				ForgeVersion:     flavorVersion,
			}
		case "paper":
			provider = &server.Paper{
				MinecraftVersion: minecraftVersion,
				PaperVersion:     flavorVersion,
			}
		case "spigot":
			provider = &server.Spigot{
				MinecraftVersion: minecraftVersion,
				SpigotVersion:    flavorVersion,
			}
		}

		err = provider.Provide(serverDir, serverFile)
		if err != nil {
			log.Fatalln(err)
		}
	},
}
