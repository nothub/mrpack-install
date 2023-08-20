package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	"github.com/nothub/mrpack-install/files"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/server"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/mrpack-install/web/download"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	// global options
	host       string
	serverDir  string
	serverFile string
	proxy      string
	dlThreads  uint8
	dlRetries  uint8
)

func init() {
	var printVersion bool
	rootCmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Print version and exit")

	var verboseLogs bool
	rootCmd.PersistentFlags().BoolVarP(&verboseLogs, "verbose", "v", false, "Enable verbose output")

	// TODO: --eula (usage: "Set this flag or MC_EULA=true to agree with Mojangs EULA: https://account.mojang.com/documents/minecraft_eula")
	// TODO: --op <uuid>...
	// TODO: --whitelist <uuid>...
	// TODO: --start-server

	rootCmd.PersistentFlags().StringVar(&host, "host", "api.modrinth.com", "Labrinth host address")
	rootCmd.PersistentFlags().StringVar(&serverDir, "server-dir", "mc", "Server directory path")
	rootCmd.PersistentFlags().StringVar(&serverFile, "server-file", "", "Server jar file name")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "Proxy url for http connections")
	rootCmd.PersistentFlags().Uint8Var(&dlThreads, "dl-threads", 8, "Concurrent download threads")
	rootCmd.PersistentFlags().Uint8Var(&dlRetries, "dl-retries", 3, "Retries when download fails")

	cobra.OnInitialize(func() {
		if printVersion {
			buildinfo.PrintInfos()
			os.Exit(0)
		}

		if verboseLogs {
			// TODO: set log level
		}

		// TODO: validate all inputs

		// --server-dir
		serverDir = strings.TrimSpace(serverDir)
		if serverDir == "" {
			log.Fatalln("invalid value for flag --server-dir")
		}
		absServerDir, err := filepath.Abs(serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		serverDir = absServerDir

		// -- server-file
		serverFile = strings.TrimSpace(serverFile)
		if serverFile != "" && serverFile != filepath.Base(serverFile) {
			log.Fatalln("invalid value for flag --server-file")
		}

		if proxy != "" {
			err := web.DefaultClient.SetProxy(proxy)
			if err != nil {
				log.Fatalln(err)
			}
		}
	})
}

var rootCmd = &cobra.Command{
	Use:   "mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>])",
	Short: "Modrinth Modpack server deployment",
	Long:  `Deploys a Modrinth modpack including Minecraft server.`,
	Example: `  mrpack-install https://example.org/data/cool-pack.mrpack
  mrpack-install downloads/cool-pack.mrpack --proxy socks5://127.0.0.1:7890
  mrpack-install adrenaserver --server-file srv.jar
  mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir mcserver
  mrpack-install communitypack9000 --host api.labrinth.example.org
  mrpack-install --version`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		index, zipPath := handleArgs(input, version, serverDir, host)

		fmt.Printf("Installing %q from %q to %q\n", index.Name, zipPath, serverDir)
		err := os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Chdir(serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		for _, file := range index.Files {
			files.AssertSafe(filepath.Join(serverDir, file.Path), serverDir)
		}

		// download server if not present
		if !files.IsFile(filepath.Join(serverDir, serverFile)) {
			fmt.Println("Server file not present, downloading...")
			fmt.Println("(Point --server-dir and --server-file to existing targets to skip this step)")
			inst := server.InstallerFromDeps(&index.Deps)
			err := inst.Install(serverDir, serverFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			fmt.Println("Server file already present, proceeding...")
		}

		// downloads
		downloads := index.ServerDownloads()
		fmt.Printf("Downloading %v dependencies...\n", len(downloads))
		downloader := download.Downloader{
			Downloads: downloads,
			Threads:   int(dlThreads),
			Retries:   int(dlRetries),
		}
		downloader.Download(serverDir)

		// overrides
		fmt.Println("Extracting overrides...")
		err = mrpack.ExtractOverrides(zipPath, serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		// save state file
		packState, err := update.BuildPackState(index, zipPath)
		if err != nil {
			log.Fatalln(err)
		}
		err = packState.Save(serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		files.RmEmptyDirs(serverDir)

		fmt.Println("Installation done :) Have a nice day ✌️")
	},
}

func handleArgs(input string, version string, serverDir string, host string) (*mrpack.Index, string) {
	err := os.MkdirAll(serverDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	archivePath := ""
	if files.IsFile(input) {
		archivePath = input

	} else if web.IsValidHttpUrl(input) {
		fmt.Println("Downloading mrpack file from", input)
		file, err := web.DefaultClient.DownloadFile(input, serverDir, "")
		if err != nil {
			log.Fatalln(err.Error())
		}
		archivePath = file

	} else {
		// input is project id or slug
		versions, err := modrinth.NewClient(host).GetProjectVersions(input, nil)
		if err != nil {
			log.Fatalln(err)
		}

		var fileInfos []modrinth.File = nil
		for i := range versions {
			if version != "" {
				if versions[i].VersionNumber == version {
					fileInfos = versions[i].Files
					break
				}
			} else {
				// latest stable release if version not specified
				if versions[i].VersionType == modrinth.ReleaseVersionType {
					fileInfos = versions[i].Files
					break
				}
			}
		}
		if len(fileInfos) == 0 {
			log.Fatalln("No files found for", input, version)
		}

		for i := range fileInfos {
			if strings.HasSuffix(fileInfos[i].Filename, ".mrpack") {
				fmt.Println("Downloading mrpack file from", fileInfos[i].Url)
				file, err := web.DefaultClient.DownloadFile(fileInfos[i].Url, serverDir, "")
				if err != nil {
					log.Fatalln(err.Error())
				}
				archivePath = file
				break
			}
		}

		if archivePath == "" {
			log.Fatalln("No mrpack file found for", input, version)
		}
	}

	if archivePath == "" {
		log.Fatalln("Unable to handle input: ", input, version)
	}

	index, err := mrpack.ReadIndex(archivePath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return index, archivePath
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}
