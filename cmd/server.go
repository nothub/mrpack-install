package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/http"
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().String("loader-version", "latest", "Mod loader version")
	/*
	   TODO: eula flag
	   TODO: ops flag
	   TODO: whitelist flag
	*/
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:       "server (vanilla|fabric|forge|quilt|paper|spigot)",
	Short:     "Prepare a server environment",
	Long:      `Download and configure one of several Minecraft server flavors.`,
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

		if minecraftVersion == "" || minecraftVersion == "latest" {
			latestMinecraftVersion, err := mojang.LatestVersion()
			if err != nil {
				log.Fatalln(err)
			}
			minecraftVersion = latestMinecraftVersion
		}

		var supplier server.DownloadSupplier = nil
		switch args[0] {
		case "vanilla":
			log.Fatalln("Not yet implemented!")
		case "fabric":
			supplier = &server.Fabric{
				MinecraftVersion: minecraftVersion,
				FabricVersion:    loaderVersion,
			}
		case "forge":
			log.Fatalln("Not yet implemented!")
		case "quilt":
			log.Fatalln("Not yet implemented!")
		case "paper":
			supplier = &server.Paper{
				MinecraftVersion: minecraftVersion,
				PaperVersion:     loaderVersion,
			}
		case "spigot":
			log.Fatalln("Not yet implemented!")
		}

		url, err := supplier.GetUrl()
		if err != nil {
			log.Fatalln(err)
		}

		file, err := http.Instance.DownloadFile(url, ".")
		if err != nil {
			return
		}
		fmt.Println("Server jar downloaded to:", file)
	},
}
