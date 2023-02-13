package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/buildinfo"
	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/http"
	"github.com/nothub/mrpack-install/http/download"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/server"
	"github.com/nothub/mrpack-install/update"
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
	rootCmd.PersistentFlags().String("host", "api.modrinth.com", "Labrinth host address")
	rootCmd.PersistentFlags().String("server-dir", "mc", "Server directory path")
	rootCmd.PersistentFlags().String("server-file", "", "Server jar file name")
	rootCmd.PersistentFlags().String("proxy", "", "Proxy url for http connections")
	rootCmd.PersistentFlags().Int("dl-threads", 8, "Concurrent download threads")
	rootCmd.PersistentFlags().Int("dl-retries", 3, "Retries when download fails")
}

type GlobalOpts struct {
	Host       string
	ServerDir  string
	ServerFile string
	Proxy      string
	DlThreads  int
	DlRetries  int
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
	serverDir, err = filepath.Abs(serverDir)
	if err != nil {
		log.Fatalln(err)
	}
	opts.ServerDir = serverDir

	serverFile, err := cmd.Flags().GetString("server-file")
	if err != nil {
		log.Fatalln(err)
	}
	serverFile, err = filepath.Abs(serverFile)
	if err != nil {
		log.Fatalln(err)
	}
	opts.ServerFile = serverFile

	proxy, err := cmd.Flags().GetString("proxy")
	if err != nil {
		log.Fatalln(err)
	}
	if proxy != "" {
		err := http.DefaultClient.SetProxy(proxy)
		if err != nil {
			log.Fatalln(err)
		}
	}
	opts.Proxy = proxy

	dlThreads, err := cmd.Flags().GetInt("dl-threads")
	if err != nil || dlThreads > 64 {
		dlThreads = 8
		fmt.Println(err)
	}
	opts.DlThreads = dlThreads

	retryTimes, err := cmd.Flags().GetInt("dl-retries")
	if err != nil {
		retryTimes = 3
		fmt.Println(err)
	}
	opts.DlRetries = retryTimes

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
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		opts := GetRootOpts(cmd)

		if opts.Version {
			fmt.Println("mrpack-install", buildinfo.Version)
			return
		}

		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		index, zipPath := handleArgs(input, version, opts.ServerDir, opts.Host)

		fmt.Printf("Installing %q from %q to %q", index.Name, zipPath, opts.ServerDir)

		for _, file := range index.Files {
			files.AssertSafe(file.Path, opts.ServerDir)
		}

		// download server if not present
		if !files.IsFile(path.Join(opts.ServerDir, opts.ServerFile)) {
			fmt.Println("Server file not present, downloading...\n(Point --server-dir and --server-file flags to an existing server file to skip this step.)")
			inst := server.InstallerFromDeps(&index.Deps)
			err := inst.Install(opts.ServerDir, opts.ServerFile)
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
			Threads:   opts.DlThreads,
			Retries:   opts.DlRetries,
		}
		downloader.Download(opts.ServerDir)

		// overrides
		fmt.Println("Extracting overrides...")
		err := mrpack.ExtractOverrides(zipPath, opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		// save state file
		packState, err := update.BuildPackState(index, zipPath)
		if err != nil {
			log.Fatalln(err)
		}
		err = packState.Save(opts.ServerDir)
		if err != nil {
			log.Fatalln(err)
		}

		files.RmEmptyDirs(opts.ServerDir)

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

	} else if http.IsValidHttpUrl(input) {
		fmt.Println("Downloading mrpack file from", input)
		file, err := http.DefaultClient.DownloadFile(input, serverDir, "")
		if err != nil {
			log.Fatalln(err.Error())
		}
		archivePath = file

		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Println(err.Error())
			}
		}(archivePath)

	} else {
		// input is project id or slug?
		versions, err := modrinth.NewClient(host).GetProjectVersions(input, nil)
		if err != nil {
			log.Fatalln(err)
		}

		var files []modrinth.File = nil
		for i := range versions {
			if version != "" {
				if versions[i].VersionNumber == version {
					files = versions[i].Files
					break
				}
			} else {
				// latest stable release if version not specified
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
				file, err := http.DefaultClient.DownloadFile(files[i].Url, serverDir, "")
				if err != nil {
					// TODO: check next file on failure
					log.Fatalln(err.Error())
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
				fmt.Println(err.Error())
			}
		}(archivePath)
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
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
