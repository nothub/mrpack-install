# mrpack-install
[![downloads](https://img.shields.io/github/downloads/nothub/mrpack-install/total.svg?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://github.com/nothub/mrpack-install/releases/latest)
[![discord](https://img.shields.io/discord/1149744662131777546?style=flat-square&labelColor=5c5c5c&color=007D9C)](https://discord.gg/QNbTeGHBRm)
[![go pkg](https://pkg.go.dev/badge/github.com/nothub/mrpack-install.svg)](https://pkg.go.dev/github.com/nothub/mrpack-install)

A cli application for installing Minecraft servers and [Modrinth](https://modrinth.com/) [modpacks](https://docs.modrinth.com/docs/modpacks/format_definition/).

---
## Usage
#### modpack deployment
```
Deploys a Modrinth modpack including Minecraft server.

Usage:
  mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>]) [flags]
  mrpack-install [command]

Examples:
  mrpack-install https://example.org/data/cool-pack.mrpack
  mrpack-install downloads/cool-pack.mrpack --proxy socks5://127.0.0.1:7890
  mrpack-install adrenaserver --server-file srv.jar
  mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir mcserver
  mrpack-install communitypack9000 --host api.labrinth.example.org
  mrpack-install --version

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  ping        Ping a Labrinth instance
  server      Prepare a plain server environment
  update      Update the deployed modpack
  version     Print version infos

Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
  -h, --help                 help for mrpack-install
      --host string          Labrinth host address (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Server jar file name
  -v, --verbose              Enable verbose output
  -V, --version              Print version and exit

Use "mrpack-install [command] --help" for more information about a command.
```
#### modpack update
```
Update the deployed modpacks files, creating backups if necessary.

Usage:
  mrpack-install update [<version>] [flags]

Flags:
      --backup-dir string   Backup directory path
  -h, --help                help for update

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Server jar file name
  -v, --verbose              Enable verbose output
```
#### plain server deployment
```
Download and configure one of several Minecraft server flavors.

Usage:
  mrpack-install server ( vanilla | fabric | quilt | forge | neoforge | paper ) [flags]

Examples:
  mrpack-install server fabric --server-dir fabric-srv
  mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar

Flags:
      --flavor-version string      Flavor version (default "latest")
  -h, --help                       help for server
      --minecraft-version string   Minecraft version (default "latest")

Global Flags:
      --dl-retries uint8     Retries when download fails (default 3)
      --dl-threads uint8     Concurrent download threads (default 8)
      --host string          Labrinth host address (default "api.modrinth.com")
      --proxy string         Proxy url for http connections
      --server-dir string    Server directory path (default "mc")
      --server-file string   Server jar file name
  -v, --verbose              Enable verbose output
```
## Install
### Linux
```sh
# download
curl -sSL -o "/tmp/mrpack-install" "https://github.com/nothub/mrpack-install/releases/download/v0.16.4/mrpack-install-linux"
# install to a place in PATH
sudo install -t "/usr/local/bin" "/tmp/mrpack-install"
# run
mrpack-install --help
```
