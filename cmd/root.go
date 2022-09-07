package cmd

import (
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/server"
	"github.com/nothub/mrpack-install/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func init() {
	// TODO: rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version infos")
	// TODO: rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().String("host", "api.modrinth.com", "Labrinth host")
	rootCmd.Flags().String("server-dir", "mc", "Server directory path")
	rootCmd.Flags().String("server-file", "", "Server jar file name")
	rootCmd.Flags().String("proxy", "", "Use a proxy to download")
	rootCmd.Flags().Int("download-threads", 8, "Download threads")
	rootCmd.Flags().Int("retry-times", 3, "Number of retries when a download fails")
}

var rootCmd = &cobra.Command{
	Use:   "mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>])",
	Short: "Installs Minecraft servers and Modrinth modpacks",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		host, err := cmd.Flags().GetString("host")
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
		downloadThreads, err := cmd.Flags().GetInt("download-threads")
		if err != nil || downloadThreads > 64 {
			downloadThreads = 8
			log.Println(err)
		}
		retryTimes, err := cmd.Flags().GetInt("retry-times")
		if err != nil {
			retryTimes = 3
			log.Println(err)
		}

		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}

		err = os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}

		archivePath := ""
		if util.PathIsFile(input) {
			archivePath = input

		} else if util.IsValidUrl(input) {
			log.Println("Downloading mrpack file from", args)
			file, err := requester.DefaultHttpClient.DownloadFile(input, serverDir, "")
			if err != nil {
				log.Fatalln(err)
			}
			archivePath = file

		} else { // input is project id or slug?
			versions, err := modrinth.NewClient(host).GetProjectVersions(input, nil)
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
					log.Println("Downloading mrpack file from", files[i].Url)
					file, err := requester.DefaultHttpClient.DownloadFile(files[i].Url, serverDir, "")
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
		}

		if archivePath == "" {
			log.Fatalln("An error occured!")
		}

		log.Println("Processing mrpack file", archivePath)

		index, err := mrpack.ReadIndex(archivePath)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Installing:", index.Name)
		log.Printf("Flavor dependencies: %+v\n", index.Dependencies)

		// download server if not present
		if !util.PathIsFile(path.Join(serverDir, serverFile)) {
			log.Println("Server file not present, downloading...")
			log.Println("(Point --server-dir and --server-file flags for an existing server file to skip this step.)")

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

			err = provider.Provide(serverDir, serverFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Server file already present, proceeding...")
		}

		// mod downloads
		log.Printf("Downloading %v dependencies...\n", len(index.Files))
		var downloads []*requester.Download
		for i := range index.Files {
			file := index.Files[i]
			if file.Env.Server == modrinth.UnsupportedEnvSupport {
				continue
			}
			downloads = append(downloads, requester.NewDownload(file.Downloads, map[string]string{"sha1": string(file.Hashes.Sha1)}, filepath.Base(file.Path), path.Join(serverDir, filepath.Dir(file.Path))))
		}
		downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloads, downloadThreads, retryTimes)
		downloadPools.Do()
		modsUnclean := false
		for i := range downloadPools.Downloads {
			dl := downloadPools.Downloads[i]
			if !dl.Success {
				modsUnclean = true
				log.Println("Dependency downloaded Fail:", dl.FileName)
			}
		}

		// overrides
		log.Println("Extracting overrides...")
		err = mrpack.ExtractOverrides(archivePath, serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		if modsUnclean {
			log.Println("There have been problems downloading downloading mods, you probably have to fix some dependency problems manually!")
		}
		log.Println("Done :) Have a nice day ✌️")
	},
}

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
