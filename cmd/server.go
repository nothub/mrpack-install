package cmd

import (
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
	Use:   "server",
	Short: "Prepare a server environment",
	Long:  `TODO`,

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
			supplier = &server.VanillaSupplier{
				MinecraftVersion: minecraftVersion,
			}
		case "fabric":
			supplier = &server.FabricSupplier{
				MinecraftVersion: minecraftVersion,
				FabricVersion:    loaderVersion,
			}
		case "forge":
			supplier = &server.ForgeSupplier{
				MinecraftVersion: minecraftVersion,
				ForgeVersion:     loaderVersion,
			}
		case "quilt":
			supplier = &server.QuiltSupplier{
				MinecraftVersion: minecraftVersion,
				QuiltVersion:     loaderVersion,
			}
		case "paper":
			supplier = &server.PaperSupplier{
				MinecraftVersion: minecraftVersion,
				PaperVersion:     loaderVersion,
			}
		case "spigot":
			supplier = &server.SpigotSupplier{
				MinecraftVersion: minecraftVersion,
				SpigotVersion:    loaderVersion,
			}
		}
		url, err := supplier.GetUrl()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(&url.Path)
	},
}
