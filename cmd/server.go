package cmd

import (
	"github.com/nothub/gorinth/server"
	"github.com/spf13/cobra"
	"log"
)

var mcVer *string
var loaderVer *string

func init() {
	mcVer = pingCmd.Flags().String("minecraft-version", "1.19.2", "Minecraft version")
	loaderVer = pingCmd.Flags().String("loader-version", "latest", "Mod loader version")

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
		log.Println("Installing", args[0], "server...")
		var supplier server.DownloadSupplier = nil
		switch args[0] {
		case "vanilla":
			supplier = &server.VanillaSupplier{
				MinecraftVersion: *mcVer,
			}
		case "fabric":
			supplier = &server.FabricSupplier{
				MinecraftVersion: *mcVer,
				FabricVersion:    *loaderVer,
			}
		case "forge":
			supplier = &server.ForgeSupplier{
				MinecraftVersion: *mcVer,
				ForgeVersion:     *loaderVer,
			}
		case "quilt":
			supplier = &server.QuiltSupplier{
				MinecraftVersion: *mcVer,
				QuiltVersion:     *loaderVer,
			}
		case "paper":
			supplier = &server.PaperSupplier{
				MinecraftVersion: *mcVer,
				PaperVersion:     *loaderVer,
			}
		case "spigot":
			supplier = &server.SpigotSupplier{
				MinecraftVersion: *mcVer,
				SpigotVersion:    *loaderVer,
			}
		}
		url, err := supplier.GetUrl()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(&url.Path)
	},
}
