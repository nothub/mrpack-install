<!-- Generated from tools/docs/usage.tmpl by: go run ./tools/docs -->
# Usage

mrpack-install sets up a Minecraft server from a [Modrinth](https://modrinth.com/) modpack and keeps it on the pack version you ask for. Run the same command again later with another version and it updates the server dir: new files are written, dropped ones deleted, and your own changes are backed up first.

- [Quick start](#quick-start)
- [Choosing the pack](#choosing-the-pack)
- [Updating a server](#updating-a-server)
- [Running it on every server start](#running-it-on-every-server-start)
- [Keeping your changes](#keeping-your-changes)
- [Checking a run before it happens](#checking-a-run-before-it-happens)
- [A stable start command](#a-stable-start-command)
- [Plain servers without a pack](#plain-servers-without-a-pack)
- [When something goes wrong](#when-something-goes-wrong)
- [Upgrading from older releases](#upgrading-from-older-releases)
- [Commands](#commands): every command and flag

## Quick start

Install the latest release of a pack into `srv`, accept Minecraft's EULA, and start the server:

```sh
mrpack-install adrenaserver --server-dir srv --server-file server.jar
echo 'eula=true' > srv/eula.txt
cd srv && java -jar server.jar nogui
```

mrpack-install downloads the pack, installs the Minecraft server with the loader the pack needs, and writes `srv/packstate.json`. That file records what was installed, which is what lets later runs update the dir. Leave it in place.

Without `--server-dir` the server goes into `mc` in the current directory. Without `--server-file` the start jar keeps the name the loader gives it. See [A stable start command](#a-stable-start-command).

## Choosing the pack

A pack is a Modrinth project slug or id, an `.mrpack` file, or an `.mrpack` URL. A version number after it picks that version; without one you get the latest release.

```sh
mrpack-install adrenaserver --server-dir srv                  # latest release, by slug
mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir srv           # a version, by project id
mrpack-install ./cool-pack.mrpack --server-dir srv            # a local file
mrpack-install https://example.org/data/cool-pack.mrpack --server-dir srv
```

Pack archives are cached in your user cache dir. For a slug or id, a version that is already cached isn't downloaded again.

Some packs have optional mods. All of them are installed by default. Pick some by file name (a regular expression), or turn them all off:

```sh
mrpack-install ./example.mrpack --server-dir srv --optional-select 'foo\.jar' --optional-select 'bar-[\d+\.]+\.jar'
mrpack-install ./example.mrpack --server-dir srv --optional-disable-all
```

Behind a proxy, or with a self-hosted Modrinth backend (labrinth):

```sh
mrpack-install ./cool-pack.mrpack --server-dir srv --proxy socks5://127.0.0.1:7890
mrpack-install communitypack9000 --server-dir srv --host api.labrinth.example.org
```

## Updating a server

Run the install command again with the new version:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv
```

or let `update` take the pack from `srv/packstate.json`:

```sh
mrpack-install update 2.0.0 --server-dir srv      # a version
mrpack-install update --server-dir srv            # the latest release
```

Either way, the server dir ends up at that version:

- Files the new version ships are written, and files it no longer ships are deleted.
- A file you edited since the last run is moved to `srv/backups/<timestamp>/` before it is replaced or deleted.
- Files no pack version ever installed are left alone: your world, your own mods, configs that mods create at runtime.
- When the Minecraft or loader version changes, the server is reinstalled and the old loader's jar removed.
- Everything is downloaded and checked before the server dir changes. A failed download stops the run and changes nothing.

`update` refuses to go back to a version published before the installed one. Pass `--allow-downgrade` when you mean it:

```sh
mrpack-install update 1.0.0 --server-dir srv                    # exits 1
mrpack-install update 1.0.0 --server-dir srv --allow-downgrade
```

## Running it on every server start

A run for the version that is already installed, with the same options, does nothing. For a slug or id it doesn't even download the pack. That makes it cheap to run before every server start, for example in a container's start script:

```sh
#!/bin/sh
set -eu
mrpack-install adrenaserver 2.0.0 --server-dir /srv --server-file server.jar
cd /srv && exec java -Xmx4G -jar server.jar nogui
```

With a version, the server stays on it until you change the script. Without one, every start moves to the latest release as soon as it is published.

## Keeping your changes

Edited files are backed up, not kept: the next update still replaces a config the pack ships. To keep your version of a file for good, exclude it. Excluded paths are never written, deleted or backed up:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv --exclude 'config/tuned/**' --exclude 'mods/*-client.jar'
```

Patterns are relative to the server dir. `*` matches within one path segment, and `**` matches any number of segments. Pass the same `--exclude` flags on every run.

Backups go to `backups/` in the server dir. To keep them elsewhere, for example on another volume:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv --backup-dir /backups
```

## Checking a run before it happens

`--dry-run` prints what a run would change and exits without changing anything, not even creating the server dir:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv --dry-run
```

```
delete config/dropped.txt
write config/new.txt
backup+write config/tuned.txt
install server: no
```

Each line is one file: `write`, `delete`, or `backup+write` and `backup+delete` for files you edited. The last line says whether the server would be reinstalled.

## A stable start command

Loaders name their server jar after their version, so a loader upgrade changes the file you start. `--server-file` gives it a fixed name instead, which keeps start scripts and panel settings working across upgrades:

```sh
mrpack-install adrenaserver --server-dir srv --server-file server.jar
```

Most loaders start with `java -jar`. Forge 1.17 and later and NeoForge also need their JVM arguments file:

```sh
cd srv && java -jar server.jar nogui                          # vanilla, Fabric, Quilt, Paper, Forge 1.16.5 and older
cd srv && java @user_jvm_args.txt -jar server.jar nogui       # Forge 1.17 and later, NeoForge
```

Use the same `--server-file` on every run. Changing the name, or deleting the jar, reinstalls the server on the next run. Quilt needs a name other than `server.jar`, which its launcher loads, so use for example `--server-file mc.jar` there.

## Plain servers without a pack

`server` installs only a Minecraft server, with a loader of your choice, and prints the file that starts it:

```sh
mrpack-install server fabric --minecraft-version 1.21.1 --server-dir srv
mrpack-install server quilt --minecraft-version 1.20.1 --flavor-version 0.20.1 --server-dir srv
mrpack-install server forge --minecraft-version 1.20.1 --flavor-version 47.3.0 --server-dir srv
mrpack-install server paper --minecraft-version 1.21.1 --server-dir srv --server-file server.jar
```

The output is only the start file, so a script can use it directly:

```sh
JAR=$(mrpack-install server fabric --minecraft-version 1.21.1 --server-dir srv)
cd srv && java -jar "$JAR" nogui
```

Forge 1.17 to 1.20.2 and NeoForge print `run.sh` (`run.bat` on Windows) when no `--server-file` is given, since only their start script starts them.

## When something goes wrong

**A download failed.** The run exits non-zero, names the files, and leaves the server dir as it was. Run it again.

**Files are missing or broken.** A normal re-run doesn't repair anything. `--force-sync` restores missing files, backs up and resets edited ones, and reinstalls the server:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv --force-sync
```

**An update replaced a file you needed.** Every run that displaced files has its own dir under `backups/`, named by time, with the files at their original paths. Copy the file back, and add an `--exclude` for it if the pack should leave it alone from now on:

```sh
ls srv/backups/
cp srv/backups/20261028221500/config/tuned.toml srv/config/tuned.toml
```

## Upgrading from older releases

Server dirs installed by v0.16.10 (with a `modpack.json`) or by a beta (with a schema-less `packstate.json`) update on their first run, and their state is converted. v0.16.10 didn't record the Modrinth project, so run the install command once instead of `update`:

```sh
mrpack-install adrenaserver 2.0.0 --server-dir srv
```

What changed:

- v0.16.10's `update <pack> [<version>]` is now `mrpack-install <pack> [<version>]`. `update` takes only a version.
- A plain re-run no longer repairs edited or deleted files. Use `--force-sync`.
- `--server-file` no longer skips the server install when that file already exists. The first run in a server dir always installs the server and backs up the existing file.

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
