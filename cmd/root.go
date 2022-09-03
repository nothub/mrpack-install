package cmd

import (
	"github.com/nothub/mrpack-install/http"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/server"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	// TODO: rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version infos")
	// TODO: rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().String("host", "api.modrinth.com", "Labrinth host")
	rootCmd.Flags().String("server-dir", "mc", "Server directory path")
	rootCmd.Flags().String("server-file", "", "Server jar file name")
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
			file, err := http.Instance.DownloadFile(input, serverDir, "")
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
					file, err := http.Instance.DownloadFile(files[i].Url, serverDir, "")
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
			log.Println("Downloading mrpack file from", u)
			_, err = http.Instance.DownloadFile(u, serverDir, serverFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Server jar file already present, skipping download...")
		}

		// download mods
		log.Printf("Downloading %v dependencies...\n", len(index.Files))
		for i := range index.Files {
			file := index.Files[i]
			if file.Env.Server == modrinth.UnsupportedEnvSupport {
				continue
			}
			success := false
			// TODO: run x downloads parallel in goroutine
			for j := range file.Downloads {
				f, err := http.Instance.DownloadFile(file.Downloads[j], path.Join(serverDir, filepath.Dir(file.Path)), filepath.Base(file.Path))
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Dependency downloaded:", f)
				}
				success = true
			}
			if !success {
				log.Fatalf("Unable to download dependency: %+v\n", file)
			}
		}

		log.Println("Extracting overrides...")
		err = mrpack.ExtractOverrides(archivePath, serverDir)
		if err != nil {
			log.Fatalln(err)
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
