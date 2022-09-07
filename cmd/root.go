package cmd

import (
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/server"
	"github.com/spf13/cobra"
	"log"
	"net/url"
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
	rootCmd.Flags().Int("download-thread", 8, "Download threads")
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
		downloadThreads, err := cmd.Flags().GetInt("download-thread")
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
		if isFilePath(input) {
			archivePath = input

		} else if isUrl(input) {
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

		log.Printf("Loader dependencies: %+v\n", index.Dependencies)

		// download server if not present
		if serverFile != "" && !isFilePath(path.Join(serverDir, serverFile)) {
			// Determine server platform
			var supplier server.DownloadSupplier = nil
			if index.Dependencies.Fabric != "" {
				supplier = &server.Fabric{
					MinecraftVersion: index.Dependencies.Minecraft,
					FabricVersion:    index.Dependencies.Fabric,
				}
			} else if index.Dependencies.Quilt != "" || index.Dependencies.Forge != "" {
				log.Fatalln("Automatic server deployment not yet implemented for this platform! Supply the path to an existing server jar file with the --server-dir and --server-file flags.")
			} else {
				// TODO: vanilla server download
			}

			// Download server
			u, err := supplier.GetUrl()
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Downloading server file from", u)
			_, err = requester.DefaultHttpClient.DownloadFile(u, serverDir, serverFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Server jar file already present, skipping download...")
		}

		// download mods
		log.Printf("Downloading %v dependencies...\n", len(index.Files))
		var downloadPoolArray []*requester.DownloadPool
		for i := range index.Files {
			file := index.Files[i]
			if file.Env.Server == modrinth.UnsupportedEnvSupport {
				continue
			}
			downloadPoolArray = append(downloadPoolArray, requester.NewDownloadPool(file.Downloads, map[string]string{"sha1": string(file.Hashes.Sha1)}, filepath.Base(file.Path), path.Join(serverDir, filepath.Dir(file.Path))))
		}

		downloadPools := requester.NewDownloadPools(requester.DefaultHttpClient, downloadPoolArray, downloadThreads, retryTimes)
		downloadPools.Do()
		log.Println("Extracting overrides...")
		err = mrpack.ExtractOverrides(archivePath, serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		uncleanNotification := false
		for i := range downloadPools.DownloadPool {
			dl := downloadPools.DownloadPool[i]
			if !dl.Success {
				uncleanNotification = true
				log.Println("Dependency downloaded Fail:", dl.FileName)
			}
		}
		if uncleanNotification {
			log.Fatalln("Download failed,You can fix the error manually")
		}

		log.Println("Done :) Have a nice day ✌️")
	},
}

func isFilePath(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}

func isUrl(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		return false
	}
	return true
}

func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
