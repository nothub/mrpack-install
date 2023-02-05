package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/server"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func init() {
	rootCmd.Flags().BoolP("version", "V", false, "Print version and exit")
	// TODO: --eula
	// TODO: --op <uuid>...
	// TODO: --whitelist <uuid>...
	// TODO: --start-server

	// TODO: rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("host", "api.modrinth.com", "Labrinth host")
	rootCmd.PersistentFlags().String("server-dir", "mc", "Server directory path")
	rootCmd.PersistentFlags().String("server-file", "", "Server jar file name")
	rootCmd.PersistentFlags().String("proxy", "", "Use a proxy to download")
	rootCmd.PersistentFlags().Int("download-threads", 8, "Download threads")
	rootCmd.PersistentFlags().Int("retry-times", 3, "Number of retries when a download fails")
}

type GlobalOpts struct {
	Host            string
	ServerDir       string
	ServerFile      string
	Proxy           string
	DownloadThreads int
	RetryTimes      int
}

func GlobalOptions(cmd *cobra.Command) *GlobalOpts {
	var opts GlobalOpts

	// TODO: validate inputs

	host, err := cmd.Flags().GetString("host")
	if err != nil {
		log.Fatalln(err)
	}
	opts.Host = host

	serverDir, err := cmd.Flags().GetString("server-dir")
	if err != nil {
		log.Fatalln(err)
	}
	opts.ServerDir = serverDir

	serverFile, err := cmd.Flags().GetString("server-file")
	if err != nil {
		log.Fatalln(err)
	}
	opts.ServerFile = serverFile

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
	opts.Proxy = proxy

	downloadThreads, err := cmd.Flags().GetInt("download-threads")
	if err != nil || downloadThreads > 64 {
		downloadThreads = 8
		fmt.Println(err)
	}
	opts.DownloadThreads = downloadThreads

	retryTimes, err := cmd.Flags().GetInt("retry-times")
	if err != nil {
		retryTimes = 3
		fmt.Println(err)
	}
	opts.RetryTimes = retryTimes

	return &opts
}

type RootOpts struct {
	*GlobalOpts
	Version bool
}

func GetRootOpts(cmd *cobra.Command) *RootOpts {
	var opts RootOpts
	opts.GlobalOpts = GlobalOptions(cmd)

	version, err := cmd.Flags().GetBool("version")
	if err != nil {
		log.Fatalln(err)
	}
	opts.Version = version

	return &opts
}

var rootCmd = &cobra.Command{
	Use:   "mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>])",
	Short: "Modrinth Modpack server deployment",
	Long:  `Deploys a Modrinth modpack including Minecraft server.`,
	Example: `  mrpack-install https://example.org/data/cool-pack.mrpack
  mrpack-install downloads/cool-pack.mrpack --proxy socks5://127.0.0.1:7890
  mrpack-install hexmc-modpack --server-file server.jar
  mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir mcserver
  mrpack-install communitypack9000 --host api.labrinth.example.org
  mrpack-install --version`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		opts := GetRootOpts(cmd)

		ver, err := cmd.Flags().GetBool("version")
		if err != nil {
			log.Fatalln(err)
		}
		if ver {
			fmt.Println("mrpack-install", buildinfo.Version)
			return
		}

		if len(args) < 1 {
			err = cmd.Help()
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(1)
		}
		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}

		err = os.MkdirAll(opts.ServerDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}

		archivePath := ""
		if util.PathIsFile(input) {
			archivePath = input

		} else if util.IsValidUrl(input) {
			fmt.Println("Downloading mrpack file from", args)
			file, err := requester.DefaultHttpClient.DownloadFile(input, opts.ServerDir, "")
			if err != nil {
				log.Fatalln(err)
			}
			archivePath = file
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					fmt.Println(err)
				}
			}(archivePath)

		} else { // input is project id or slug?
			versions, err := modrinth.NewClient(opts.Host).GetProjectVersions(input, nil)
			if err != nil {
				log.Fatalln(err)
			}

			// get files uploaded for specified version or latest stable if not specified
			var files []modrinth.File = nil
			for i := range versions {
				if version != "" {
					if versions[i].VersionNumber == version {
						files = versions[i].Files
						break
					}
				} else {
					if versions[i].VersionType == modrinth.ReleaseVersionType {
						files = versions[i].Files
						break
					}
				}
			}
			if len(files) == 0 {
				log.Fatalln("No files found for", input, version)
			}

			for i := range files {
				if strings.HasSuffix(files[i].Filename, ".mrpack") {
					fmt.Println("Downloading mrpack file from", files[i].Url)
					file, err := requester.DefaultHttpClient.DownloadFile(files[i].Url, opts.ServerDir, "")
					if err != nil {
						log.Fatalln(err)
					}
					archivePath = file
					break
				}
			}
			if archivePath == "" {
				log.Fatalln("No mrpack file found for", input, version)
			}
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					fmt.Println(err)
				}
			}(archivePath)
		}

		if archivePath == "" {
			log.Fatalln("An error occured!")
		}

		fmt.Println("Processing mrpack file", archivePath)

		index, err := mrpack.ReadIndex(archivePath)
		if err != nil {
			log.Fatalln(err)
		}

		for _, file := range index.Files {
			ok, err := util.PathIsSubpath(file.Path, opts.ServerDir)
			if err != nil {
				log.Println(err.Error())
			}
			if err != nil || !ok {
				log.Fatalln("File path is not safe: " + file.Path)
			}
		}

		fmt.Println("Installing:", index.Name)
		fmt.Printf("Flavor dependencies: %+v\n", index.Dependencies)

		// download server if not present
		if !util.PathIsFile(path.Join(opts.ServerDir, opts.ServerFile)) {
			fmt.Println("Server file not present, downloading...\n(Point --server-dir and --server-file flags to an existing server file to skip this step.)")

			var provider server.Provider
			if index.Dependencies.Fabric != "" {
				provider, err = server.NewProvider("fabric", index.Dependencies.Minecraft, index.Dependencies.Fabric)
				if err != nil {
					log.Fatalln(err)
				}
			} else if index.Dependencies.Quilt != "" {
				provider, err = server.NewProvider("quilt", index.Dependencies.Minecraft, index.Dependencies.Quilt)
				if err != nil {
					log.Fatalln(err)
				}
			} else if index.Dependencies.Forge != "" {
				provider, err = server.NewProvider("forge", index.Dependencies.Minecraft, index.Dependencies.Forge)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				provider, err = server.NewProvider("vanilla", index.Dependencies.Minecraft, "")
				if err != nil {
					log.Fatalln(err)
				}
			}

			err = provider.Provide(opts.ServerDir, opts.ServerFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			fmt.Println("Server file already present, proceeding...")
		}

		// mod downloads
		fmt.Printf("Downloading %v dependencies...\n", len(index.Files))
		var downloads []*requester.Download
		for i := range index.Files {
			file := index.Files[i]
			if file.Env.Server == modrinth.UnsupportedEnvSupport {
				continue
			}
			downloads = append(downloads, requester.NewDownload(file.Downloads, map[string]string{"sha1": string(file.Hashes.Sha1)}, filepath.Base(file.Path), path.Join(opts.ServerDir, filepath.Dir(file.Path))))
		}
		downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, opts.DownloadThreads, opts.RetryTimes)
		downloadPools.Do()
		modsUnclean := false
		for i := range downloadPools.Downloads {
			dl := downloadPools.Downloads[i]
			if !dl.Success {
				modsUnclean = true
				fmt.Println("Dependency downloaded Fail:", dl.FileName)
			}
		}

		// overrides
		fmt.Println("Extracting overrides...")
		err = mrpack.ExtractOverrides(archivePath, opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		info, err := update.GenerateModPackInfo(index)
		if err != nil {
			fmt.Println(err)
		}
		err = info.Write(path.Join(opts.ServerDir, "modpack.json"))
		if err != nil {
			fmt.Println(err)
		}

		if modsUnclean {
			fmt.Println("There have been problems downloading downloading mods, you probably have to fix some dependency problems manually!")
		}
		fmt.Println("Done :) Have a nice day ✌️")
	},
}

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
