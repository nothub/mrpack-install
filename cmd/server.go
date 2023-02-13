package cmd

import (
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	serverCmd.Flags().String("minecraft-version", "latest", "Minecraft version")
	serverCmd.Flags().String("flavor-version", "latest", "Flavor version")
	// TODO: --eula
	// TODO: --op <uuid>...
	// TODO: --whitelist <uuid>...
	// TODO: --start-server

	rootCmd.AddCommand(serverCmd)
}

type ServerOpts struct {
	*GlobalOpts
	MinecraftVersion string
	FlavorVersion    string
}

func GetServerOpts(cmd *cobra.Command) *ServerOpts {
	var opts ServerOpts
	opts.GlobalOpts = GlobalOptions(cmd)

	minecraftVersion, err := cmd.Flags().GetString("minecraft-version")
	if err != nil {
		log.Fatalln(err)
	}
	if minecraftVersion == "" || minecraftVersion == "latest" {
		latestMinecraftVersion, err := mojang.LatestRelease()
		if err != nil {
			log.Fatalln(err)
		}
		minecraftVersion = latestMinecraftVersion
	}
	opts.MinecraftVersion = minecraftVersion

	flavorVersion, err := cmd.Flags().GetString("flavor-version")
	if err != nil {
		log.Fatalln(err)
	}
	opts.FlavorVersion = flavorVersion

	return &opts
}

var serverCmd = &cobra.Command{
	Use:   "server (vanilla | fabric | quilt | forge | paper)",
	Short: "Prepare a plain server environment",
	Long:  `Download and configure one of several Minecraft server flavors.`,
	Example: `  mrpack-install server fabric --server-dir fabric-srv
  mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{
		server.Vanilla.String(),
		server.Fabric.String(),
		server.Quilt.String(),
		server.Forge.String(),
		server.Paper.String(),
	},
	Run: func(cmd *cobra.Command, args []string) {
		opts := GetServerOpts(cmd)

		err := os.MkdirAll(opts.ServerDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Chdir(opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		flavor := server.GetFlavor(args[0])
		inst, err := server.NewInstaller(flavor, opts.MinecraftVersion, opts.FlavorVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = inst.Install(opts.ServerDir, opts.ServerFile)
		if err != nil {
			log.Fatalln(err)
		}
	},
}
