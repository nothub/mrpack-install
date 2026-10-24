package cmd

import (
	"github.com/nothub/mrpack-install/files"
	"github.com/nothub/mrpack-install/internal/buildinfo"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/modrinth/mrpack"
	"github.com/nothub/mrpack-install/server"
	"github.com/nothub/mrpack-install/update"
	"github.com/nothub/mrpack-install/update/backup"
	"github.com/nothub/mrpack-install/update/packstate"
	"github.com/nothub/mrpack-install/web"
	"github.com/nothub/mrpack-install/web/download"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

var (
	// local options
	optionalSelected   []string
	optionalDisableAll bool
	backupDir          string
)

func init() {
	var printVersion bool
	RootCmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Print version and exit")

	var verboseLogs bool
	RootCmd.PersistentFlags().BoolVarP(&verboseLogs, "verbose", "v", false, "Enable verbose output")

	// TODO: --eula (usage: "Set this flag or MC_EULA=true to agree with Mojangs EULA: https://account.mojang.com/documents/minecraft_eula")
	// TODO: --op <uuid>...
	// TODO: --whitelist <uuid>...
	// TODO: --start-server

	RootCmd.PersistentFlags().StringVar(&host, "host", "api.modrinth.com", "Labrinth host address or full URL, e.g. http://labrinth.lan:8000")
	RootCmd.PersistentFlags().StringVar(&serverDir, "server-dir", "mc", "Server directory path")
	RootCmd.PersistentFlags().StringVar(&serverFile, "server-file", "", "Server jar file name")
	RootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "Proxy url for http connections")
	RootCmd.PersistentFlags().Uint8Var(&dlThreads, "dl-threads", 8, "Concurrent download threads")
	RootCmd.PersistentFlags().Uint8Var(&dlRetries, "dl-retries", 3, "Retries when download fails")

	RootCmd.Flags().StringArrayVar(&optionalSelected, "optional-select", nil, "Select optional mods by file path (regex)")
	RootCmd.Flags().BoolVar(&optionalDisableAll, "optional-disable-all", false, "Disable all optional mods")
	RootCmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory path (default <server-dir>/backups)")

	cobra.OnInitialize(func() {
		if printVersion {
			buildinfo.Print()
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

		if runtime.GOOS == "windows" {
			// normalize lower case windows path
			// e.g. c:\MCServers to C:\MCServers
			serverDir = strings.ToUpper(serverDir[0:1]) + serverDir[1:]
		}

		// --backup-dir, before the run changes into the server dir
		if backupDir != "" {
			absBackupDir, err := filepath.Abs(backupDir)
			if err != nil {
				log.Fatalln(err)
			}
			backupDir = absBackupDir
		}

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

		modrinth.Client = modrinth.NewClient(host)
	})
}

var RootCmd = &cobra.Command{
	Use:   "mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>])",
	Short: "Modrinth Modpack server deployment",
	Long:  `Deploys a Modrinth modpack including Minecraft server.`,
	Example: `  mrpack-install https://example.org/data/cool-pack.mrpack
  mrpack-install downloads/cool-pack.mrpack --proxy socks5://127.0.0.1:7890
  mrpack-install adrenaserver --server-file srv.jar
  mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir mcserver
  mrpack-install communitypack9000 --host api.labrinth.example.org
  mrpack-install example.mrpack --optional-select 'foo\.jar' \
                                --optional-select 'bar-[\d+\.]+\.jar'`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		index, zipPath, lookup := handleArgs(input, version, serverDir, host)
		opts := packstate.Options{OptionalSelect: optionalSelected, OptionalDisableAll: optionalDisableAll}

		zipPath, err := filepath.Abs(zipPath)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Installing %q from %q to %q\n", index.Name, zipPath, serverDir)
		err = os.MkdirAll(serverDir, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Chdir(serverDir)
		if err != nil {
			log.Fatalln(err)
		}

		old, err := packstate.LoadPackState(serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		plan, err := update.Compute(serverDir, old, index, zipPath, opts)
		if err != nil {
			log.Fatalln(err)
		}

		// download server if not present
		if !files.IsFile(filepath.Join(serverDir, serverFile)) {
			log.Println("Server file not present, downloading...")
			log.Println("(Point --server-dir and --server-file to existing targets to skip this step)")
			inst := server.InstallerFromDeps(&index.Deps)
			err := inst.Install(serverDir, serverFile)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Println("Server file already present, proceeding...")
		}

		packState, err := packstate.FromArchive(zipPath, opts, lookup)
		if err != nil {
			log.Fatalln(err)
		}
		backupBase := backupDir
		if backupBase == "" {
			backupBase = filepath.Join(serverDir, "backups")
		}
		err = update.Apply(serverDir, plan, packState, index, zipPath, download.Downloader{
			Threads: int(dlThreads),
			Retries: int(dlRetries),
		}, backup.NewDir(backupBase))
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Installation done :) Have a nice day ✌️")
	},
}

// handleArgs resolves the pack input to its index and archive path, and
// returns how to find the pack's Modrinth ids: the ones a slug or id input
// resolved, or a lookup by archive hash for a file or URL input.
func handleArgs(input string, version string, serverDir string, host string) (*mrpack.Index, string, func(string) (packstate.Ids, error)) {
	err := os.MkdirAll(serverDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	archivePath := ""
	lookup := lookupIds
	if web.IsValidHttpUrl(input) {
		log.Printf("Downloading mrpack file from: %s\n", input)
		file, err := web.DefaultClient.DownloadFile(input, serverDir, "")
		if err != nil {
			log.Fatalln(err.Error())
		}
		archivePath = file

	} else if files.IsFile(input) {
		archivePath = input
		if runtime.GOOS == "windows" {
			archivePath = strings.ToUpper(archivePath[0:1]) + archivePath[1:]
		}
		log.Printf("Installing from mrpack file: %s\n", archivePath)

	} else {
		log.Printf("Trying to resolve project id or slug: %s\n", input)

		project, err := modrinth.Client.GetProject(input)
		if err != nil {
			log.Fatalln(err)
		}
		versions, err := modrinth.Client.GetProjectVersions(input, nil)
		if err != nil {
			log.Fatalln(err)
		}

		var picked *modrinth.Version
		for i := range versions {
			if version != "" {
				if versions[i].VersionNumber == version {
					picked = versions[i]
					break
				}
			} else {
				// latest stable release if version not specified
				if versions[i].VersionType == modrinth.ReleaseVersionType {
					picked = versions[i]
					break
				}
			}
		}
		if picked == nil || len(picked.Files) == 0 {
			log.Fatalln("No files found for", input, version)
		}
		fileInfos := picked.Files
		ids := packstate.Ids{Slug: project.Slug, ProjectId: project.Id, VersionId: picked.Id}
		lookup = func(string) (packstate.Ids, error) { return ids, nil }

		for i := range fileInfos {
			if strings.HasSuffix(fileInfos[i].Filename, ".mrpack") {
				log.Println("Downloading mrpack file from", fileInfos[i].Url)
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

	return index, archivePath, lookup
}

// lookupIds finds the Modrinth ids of the archive with the given SHA-512.
func lookupIds(sha512 string) (packstate.Ids, error) {
	version, err := modrinth.Client.VersionFromHash(sha512, "sha512")
	if err != nil {
		return packstate.Ids{}, err
	}
	project, err := modrinth.Client.GetProject(version.ProjectId)
	if err != nil {
		return packstate.Ids{}, err
	}
	return packstate.Ids{Slug: project.Slug, ProjectId: project.Id, VersionId: version.Id}, nil
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}
