package cmd

import (
	"fmt"
	"github.com/nothub/mrpack-install/cache"
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
	excludes           []string
	forceSync          bool
	dryRun             bool
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
	RootCmd.PersistentFlags().StringVar(&serverFile, "server-file", "", "Name of the jar that starts the server")
	RootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "Proxy url for http connections")
	RootCmd.PersistentFlags().Uint8Var(&dlThreads, "dl-threads", 8, "Concurrent download threads")
	RootCmd.PersistentFlags().Uint8Var(&dlRetries, "dl-retries", 3, "Retries when download fails")

	addSyncFlags(RootCmd)

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

		// --exclude
		for _, p := range excludes {
			if err := packstate.CheckPattern(p); err != nil {
				log.Fatalln(err)
			}
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

// addSyncFlags registers the flags of a sync, which the root command and
// update share.
func addSyncFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.StringArrayVar(&optionalSelected, "optional-select", nil, "Select optional mods by file path (regex)")
	fs.BoolVar(&optionalDisableAll, "optional-disable-all", false, "Disable all optional mods")
	fs.StringVar(&backupDir, "backup-dir", "", "Backup directory path (default <server-dir>/backups)")
	fs.StringArrayVar(&excludes, "exclude", nil, "Never touch paths matching this pattern, relative to the server dir (repeatable, e.g. 'config/**', 'mods/*-client.jar')")
	fs.BoolVar(&forceSync, "force-sync", false, "Re-apply the pack and reinstall the server, even when it is already installed")
	fs.BoolVar(&dryRun, "dry-run", false, "Print what the run would change and exit without changing anything")
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
		old, err := packstate.LoadPackState(serverDir)
		if err != nil {
			log.Fatalln(err)
		}
		syncPack(resolvePack(input, version), old)
	},
}

// syncPack brings the server dir to the pack src, from the old state, nil
// when there is none.
func syncPack(src packSource, old *packstate.Schema) {
	opts := packstate.Options{OptionalSelect: optionalSelected, OptionalDisableAll: optionalDisableAll, Exclude: excludes, ServerFile: serverFile, ForceSync: forceSync}
	installed := func(sum string) bool {
		noop, err := update.NoOp(serverDir, old, sum, opts)
		if err != nil {
			log.Fatalln(err)
		}
		if noop {
			log.Printf("%q %s is already installed in %q, nothing to do\n", old.Name, old.Version, serverDir)
			if dryRun {
				// The empty plan, in the shape a dry-run always prints.
				fmt.Println("install server: no")
			}
		}
		return noop
	}
	// A slug or id input knows its archive's SHA-512 before downloading.
	if installed(src.sha512) {
		return
	}
	zipPath, err := filepath.Abs(src.archive())
	if err != nil {
		log.Fatalln(err)
	}
	sum, err := packstate.ArchiveSha512(zipPath)
	if err != nil {
		log.Fatalln(err)
	}
	if installed(sum) {
		return
	}
	index, err := mrpack.ReadIndex(zipPath)
	if err != nil {
		log.Fatalln(err)
	}

	plan, err := update.Compute(serverDir, old, index, zipPath, opts)
	if err != nil {
		log.Fatalln(err)
	}
	if dryRun {
		err = plan.Describe(os.Stdout)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	log.Printf("Installing %q from %q to %q\n", index.Name, zipPath, serverDir)
	err = os.MkdirAll(serverDir, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	// The Forge and Quilt installers and override extraction resolve
	// paths against the working directory.
	err = os.Chdir(serverDir)
	if err != nil {
		log.Fatalln(err)
	}

	// Before staging: its API lookup can fail, and should fail before the downloads.
	packState, err := packstate.FromArchive(zipPath, opts, src.lookup)
	if err != nil {
		log.Fatalln(err)
	}

	staging, err := update.Stage(serverDir, plan, index, zipPath, download.Downloader{
		Threads: int(dlThreads),
		Retries: int(dlRetries),
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Everything from here on changes the server dir.
	err = old.Invalidate(serverDir)
	if err != nil {
		log.Fatalln(err)
	}

	backupBase := backupDir
	if backupBase == "" {
		backupBase = filepath.Join(serverDir, "backups")
	}
	backups := backup.NewDir(backupBase)
	err = update.InstallServer(serverDir, plan, old, packState, server.InstallerFromDeps(&index.Deps), serverFile, backups)
	if err != nil {
		log.Fatalln(err)
	}

	err = update.Apply(serverDir, plan, staging, packState, zipPath, backups)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Installation done :) Have a nice day ✌️")
}

// packSource is a pack input resolved to where its archive comes from.
type packSource struct {
	// path is a local archive, read in place.
	path string
	// url is where the archive downloads from, into the archive cache.
	url string
	// sha512 is the archive's SHA-512 when known before the download, for
	// a slug or id input.
	sha512 string
	// lookup finds the pack's Modrinth ids: the ones a slug or id input
	// resolved, or a lookup by archive hash for a file or URL input.
	lookup func(string) (packstate.Ids, error)
}

// resolvePack resolves a file path, URL, slug or id input without
// downloading the archive.
func resolvePack(input string, version string) packSource {
	if web.IsValidHttpUrl(input) {
		return packSource{url: input, lookup: lookupIds}
	}
	if files.IsFile(input) {
		path := input
		if runtime.GOOS == "windows" {
			path = strings.ToUpper(path[0:1]) + path[1:]
		}
		log.Printf("Installing from mrpack file: %s\n", path)
		return packSource{path: path, lookup: lookupIds}
	}
	return resolveProject(input, version)
}

// resolveProject resolves a Modrinth project id or slug at version, or at
// its latest release when version is empty.
func resolveProject(input string, version string) packSource {
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
	ids := packstate.Ids{Slug: project.Slug, ProjectId: project.Id, VersionId: picked.Id}
	for _, f := range picked.Files {
		if strings.HasSuffix(f.Filename, ".mrpack") {
			return packSource{
				url:    f.Url,
				sha512: f.Hashes.Sha512,
				lookup: func(string) (packstate.Ids, error) { return ids, nil },
			}
		}
	}
	log.Fatalln("No mrpack file found for", input, version)
	return packSource{}
}

// archive returns the local path of the source's archive, downloading it
// into the archive cache unless a matching copy is cached.
func (src packSource) archive() string {
	if src.path != "" {
		return src.path
	}
	cacheDir, err := cache.Dir()
	if err != nil {
		log.Fatalln(err)
	}
	path, err := cache.Archive(cacheDir, src.url, src.sha512, web.DefaultClient.Download)
	if err != nil {
		log.Fatalln(err)
	}
	return path
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
