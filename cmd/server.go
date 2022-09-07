package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().String("loader-version", "latest", "Mod loader version")
	serverCmd.Flags().String("server-dir", "mc", "Server directory path")
	serverCmd.Flags().String("server-file", "", "Server jar file name")
	/*
	   TODO: eula flag
	   TODO: ops flag
	   TODO: whitelist flag
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
		loaderVersion, err := cmd.Flags().GetString("loader-version")
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
			err = os.MkdirAll("work/quilt", 0755)
			if err != nil {
				log.Fatalln(err)
			}
			// download https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/latest/quilt-installer-latest.jar
			// java -jar quilt-installer-latest.jar install server ${minecraftVersion} --download-server
			log.Fatalln("Not yet implemented!")
		case "paper":
			supplier = &server.Paper{
				MinecraftVersion: minecraftVersion,
				PaperVersion:     loaderVersion,
			}
		case "spigot":
			err = os.MkdirAll("work/spigot", 0755)
			if err != nil {
				log.Fatalln(err)
			}
			// download https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar
			// git config --global --unset core.autocrlf
			// java -jar BuildTools.jar --rev ${minecraftVersion}
			log.Fatalln("Not yet implemented!")
		}

		url, err := supplier.GetUrl()
		if err != nil {
			log.Fatalln(err)
		}

		file, err := requester.DefaultHttpClient.DownloadFile(url, serverDir, serverFile)
		if err != nil {
			return
		}
		fmt.Println("Server jar downloaded to:", file)
	},
}
