# mrpack-install

[![Go Reference](https://pkg.go.dev/badge/github.com/nothub/mrpack-install.svg)](https://pkg.go.dev/github.com/nothub/mrpack-install)

A cli application for installing Minecraft servers
and [Modrinth](https://modrinth.com/) [modpacks](https://docs.modrinth.com/docs/modpacks/format_definition/).

---

### Usage

```
mrpack-install (<filepath> | <url> | <slug> [<version>] | <id> [<version>])
mrpack-install server (vanilla | fabric | quilt | forge | paper | spigot)
mrpack-install ping
```

### Examples

```
# install from file
mrpack-install downloads/cool-pack.mrpack

# install from url
mrpack-install https://example.org/data/cool-pack.mrpack

# install from api
mrpack-install hexmc-modpack --server-file server.jar
mrpack-install yK0ISmKn 1.0.0-1.18 --server-dir mcserver
mrpack-install communitypack9000 --host api.labrinth.example.org

# install bare server
mrpack-install server fabric --server-dir fabricsrv
mrpack-install server paper --minecraft-version 1.18.2 --server-file srv.jar
```
