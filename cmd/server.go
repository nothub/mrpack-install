package cmd

import (
	"fmt"
	"github.com/nothub/gorinth/server"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "1.19.2", "Minecraft version")
	serverCmd.Flags().String("loader-version", "latest", "Mod loader version")
	/*
	   TODO: eula flag
	   TODO: ops flag
	   TODO: whitelist flag
	*/
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server (vanilla|fabric|forge|quilt|paper|spigot)",
	Short: "Prepare a server environment",
	Long:  `Download and configure one of several Minecraft server flavors.`,

	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"vanilla", "fabric", "forge", "quilt", "paper", "spigot"},

	Run: func(cmd *cobra.Command, args []string) {
		minecraftVersion, err := cmd.Flags().GetString("minecraft-version")
		if err != nil {
			log.Fatalln(err)
		}
		loaderVersion, err := cmd.Flags().GetString("loader-version")
		if err != nil {
			log.Fatalln(err)
		}

		var supplier server.DownloadSupplier = nil
		switch args[0] {
		case "vanilla":
			supplier = &server.Vanilla{
				MinecraftVersion: minecraftVersion,
			}
		case "fabric":
			supplier = &server.Fabric{
				MinecraftVersion: minecraftVersion,
				FabricVersion:    loaderVersion,
			}
		case "forge":
			supplier = &server.Forge{
				MinecraftVersion: minecraftVersion,
				ForgeVersion:     loaderVersion,
			}
		case "quilt":
			supplier = &server.Quilt{
				MinecraftVersion: minecraftVersion,
				QuiltVersion:     loaderVersion,
			}
		case "paper":
			supplier = &server.Paper{
				MinecraftVersion: minecraftVersion,
				PaperVersion:     loaderVersion,
			}
		case "spigot":
			supplier = &server.Spigot{
				MinecraftVersion: minecraftVersion,
				SpigotVersion:    loaderVersion,
			}
		}

		url, err := supplier.GetUrl()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(url)
	},
}
