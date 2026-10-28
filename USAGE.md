<!-- Generated from tools/docs/usage.tmpl by: go run ./tools/docs -->
# Usage

`<pack>` is a Modrinth project slug or id, an `.mrpack` file path, or an `.mrpack` URL. Every command takes `--server-dir` (default `mc`). [Commands](#commands) lists all flags.

## Install a pack

```sh
mrpack-install adrenaserver --server-dir srv                  # latest release
mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir srv           # a version, by project id
mrpack-install ./cool-pack.mrpack --server-dir srv            # a local file
mrpack-install https://example.org/data/cool-pack.mrpack      # a URL
```

Downloads go through a proxy, or come from a self-hosted Modrinth instance (labrinth):

```sh
mrpack-install ./cool-pack.mrpack --proxy socks5://127.0.0.1:7890
mrpack-install communitypack9000 --host api.labrinth.example.org
```

Optional mods are all installed by default. Select them by file name (regex), or disable them all:

```sh
mrpack-install ./example.mrpack --optional-select 'foo\.jar' --optional-select 'bar-[\d+\.]+\.jar'
mrpack-install ./example.mrpack --optional-disable-all
```

## Update a server

Run the same command for another version. It brings the server dir to that version:

```sh
mrpack-install adrenaserver 1.0.0 --server-dir srv
mrpack-install adrenaserver 2.0.0 --server-dir srv
```

- Files the new version ships are written, and files it dropped are deleted.
- Files no pack version listed, such as your own mods or the world, are never touched.
- A file you edited since the last run is moved to `srv/backups/<timestamp>/` before it is replaced or deleted.
- The server is reinstalled when the Minecraft or loader version changes, and the old loader's jar is removed.
- Every file is downloaded and checked first. A failed download exits non-zero and changes nothing.

`update` does the same for the pack the server dir records, at a version or at the latest release:

```sh
mrpack-install update 2.0.0 --server-dir srv
mrpack-install update --server-dir srv
```

`update` refuses a version published before the installed one:

```sh
mrpack-install update 1.0.0 --server-dir srv                    # exits 1
mrpack-install update 1.0.0 --server-dir srv --allow-downgrade
```

## Run it on every server start

A run for the installed version and options changes nothing, and for a slug or id downloads nothing:

```sh
mrpack-install adrenaserver --server-dir srv
```

It doesn't repair either. To restore deleted files, back up and reset edited ones, and reinstall the server:

```sh
mrpack-install adrenaserver --server-dir srv --force-sync
```

## Keep your changes

Never touch some paths, even when the pack changes or drops them. Patterns are relative to the server dir, and `**` matches any number of path segments:

```sh
mrpack-install ./pack.mrpack --server-dir srv --exclude 'config/tuned/**' --exclude 'mods/*-client.jar'
```

Keep backups somewhere else:

```sh
mrpack-install ./pack.mrpack --server-dir srv --backup-dir /backups
```

## See what a run would do

`--dry-run` prints one line per file it would write, delete or back up, and whether it would install the server. It changes nothing, and doesn't even create the server dir:

```sh
mrpack-install ./pack-v2.mrpack --server-dir srv --dry-run
```

```
delete config/dropped.txt
write config/new.txt
install server: no
```

## Name the start jar

`--server-file` names the jar that starts the server, and keeps that name across loader upgrades. A deleted jar or a new name triggers a reinstall:

```sh
mrpack-install ./pack.mrpack --server-dir srv --server-file mc.jar
```

Forge 1.17+ and NeoForge jars don't read `user_jvm_args.txt`:

```sh
cd srv && java @user_jvm_args.txt -jar mc.jar     # Forge 1.17+, NeoForge
cd srv && java -jar mc.jar                        # all other loaders
```

## Install a plain server

`server` installs a server without a pack and prints the file that starts it:

```sh
mrpack-install server fabric --minecraft-version 1.21.1 --server-dir srv
mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar
mrpack-install server quilt --minecraft-version 1.20.1 --flavor-version 0.20.1 --server-file mc.jar
mrpack-install server neoforge --minecraft-version 1.21.1 --server-file mc.jar
mrpack-install server forge --minecraft-version 1.16.5 --flavor-version 36.2.39 --server-file mc.jar
```

The output is the start file alone, so a script can use it:

```sh
JAR=$(mrpack-install server fabric --minecraft-version 1.21.1 --server-dir srv)
cd srv && java -jar "$JAR"
```

Forge 1.17 to 1.20.2 and NeoForge print `run.sh` (`run.bat` on Windows) when started without `--server-file`.

## Upgrading from older releases

A server dir installed by v0.16.10 (`modpack.json`) or a beta (`packstate.json`) updates on its first run, and its state is converted. A v0.16.10 dir records no Modrinth project, so pass the pack once:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv
```

Behaviour that changed:

- A plain re-run no longer repairs edited or deleted files. Use `--force-sync`.
- A server jar that is already in the server dir no longer skips the first install.
- v0.16.10's `update <pack> [<version>]` is now `mrpack-install <pack> [<version>]`. `update` takes only a version.

## Commands

### root

```
Deploys a Modrinth modpack including Minecraft server.

Examples: https://github.com/nothub/mrpack-install/blob/main/USAGE.md

Usage:
  mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>]) [flags]
  mrpack-install [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  ping        Ping a Labrinth instance
  server      Prepare a plain server environment
  update      Update the installed modpack
  version     Print version infos

Flags:
      --backup-dir string             Backup directory path (default <server-dir>/backups)
      --dl-retries uint8              Retries when download fails (default 3)
      --dl-threads uint8              Concurrent download threads (default 8)
      --dry-run                       Print what the run would change and exit without changing anything
      --exclude stringArray           Never touch paths matching this pattern, relative to the server dir (repeatable, e.g. 'config/**', 'mods/*-client.jar')
      --force-sync                    Re-apply the pack and reinstall the server, even when it is already installed
  -h, --help                          help for mrpack-install
      --host string                   Labrinth host address or full URL, e.g. http://labrinth.lan:8000 (default "api.modrinth.com")
      --optional-disable-all          Disable all optional mods
      --optional-select stringArray   Select optional mods by file path (regex)
      --proxy string                  Proxy url for http connections
      --server-dir string             Server directory path (default "mc")
      --server-file string            Name of the jar that starts the server. Forge 1.17+ and NeoForge jars don't read user_jvm_args.txt, start them as java @user_jvm_args.txt -jar <file>. Older Forge and other jars start as java -jar <file>
  -v, --verbose                       Enable verbose output
  -V, --version                       Print version and exit

Use "mrpack-install [command] --help" for more information about a command.

```

### ping

```
Connect to a Labrinth instance and display basic information.

Usage:
  mrpack-install ping [flags]

Flags:
  -h, --help   help for ping

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address or full URL, e.g. http://labrinth.lan:8000 (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Name of the jar that starts the server. Forge 1.17+ and NeoForge jars don't read user_jvm_args.txt, start them as java @user_jvm_args.txt -jar <file>. Older Forge and other jars start as java -jar <file>
  -v, --verbose              Enable verbose output

```

### server

```
Download and configure one of several Minecraft server flavors.

Prints the file that starts the server on stdout: a jar to run with
java -jar, or run.sh (run.bat on Windows) where only the installer's
start script starts it.

Examples: https://github.com/nothub/mrpack-install/blob/main/USAGE.md

Usage:
  mrpack-install server ( vanilla | fabric | quilt | forge | neoforge | paper ) [flags]

Flags:
      --flavor-version string      Flavor version (default "latest")
  -h, --help                       help for server
      --minecraft-version string   Minecraft version (default "latest")

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address or full URL, e.g. http://labrinth.lan:8000 (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Name of the jar that starts the server. Forge 1.17+ and NeoForge jars don't read user_jvm_args.txt, start them as java @user_jvm_args.txt -jar <file>. Older Forge and other jars start as java -jar <file>
  -v, --verbose              Enable verbose output

```

### update

```
Update the installed modpack to <version>, or to its latest release, the
same way mrpack-install <pack> <version> does, with the pack the server
dir's packstate.json records. A version published before the installed
one needs --allow-downgrade.

Usage:
  mrpack-install update [<version>] [flags]

Flags:
      --allow-downgrade               Update to a version published before the installed one
      --backup-dir string             Backup directory path (default <server-dir>/backups)
      --dry-run                       Print what the run would change and exit without changing anything
      --exclude stringArray           Never touch paths matching this pattern, relative to the server dir (repeatable, e.g. 'config/**', 'mods/*-client.jar')
      --force-sync                    Re-apply the pack and reinstall the server, even when it is already installed
  -h, --help                          help for update
      --optional-disable-all          Disable all optional mods
      --optional-select stringArray   Select optional mods by file path (regex)

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address or full URL, e.g. http://labrinth.lan:8000 (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Name of the jar that starts the server. Forge 1.17+ and NeoForge jars don't read user_jvm_args.txt, start them as java @user_jvm_args.txt -jar <file>. Older Forge and other jars start as java -jar <file>
  -v, --verbose              Enable verbose output

```

### version

```
Extract and display the running binaries embedded version information.

Usage:
  mrpack-install version [flags]

Flags:
  -h, --help   help for version

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address or full URL, e.g. http://labrinth.lan:8000 (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Name of the jar that starts the server. Forge 1.17+ and NeoForge jars don't read user_jvm_args.txt, start them as java @user_jvm_args.txt -jar <file>. Older Forge and other jars start as java -jar <file>
  -v, --verbose              Enable verbose output

```
