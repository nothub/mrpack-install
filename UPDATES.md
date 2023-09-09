# Updating an existing modpack

## Behaviour

TODO: use clear terminology
- `packstate`
- `old packstate`
- `new packstate`
- `packstate files/hashes`

For an updated to be executed, the program performs the following steps:

1. Read state information of old modpack version from `packstate.json` file.
    * Download the mrpack archive of the new modpack version if required.
2. Generate state information of new modpack version from mrpack archive file.
3. If any version difference of dependencies exists, cancel the update.
    * (`minecraft`, `fabric-loader`, `quilt-loader`, `forge`, `neoforge`)
    * TODO: flag for continuing anyway
    * TODO: actually also keep track of and update the dependencies
4. For all files not changed by the update:
    * If the file is stored inside the `mods` directory and
    * the file is present in the old modpack version and
    * the file is not present in the new modpack version, then
    * create a backup of the file and delete the file.
5. For all files changed by the update:
    * If the file is not stored inside the `mods` directory and
    * the file hash is not equal to the hash in the old modpack version and
    * the file hash is not equal to the hash in the new modpack version, then
    * create a backup of the file and overwrite the file.

TODO: is this actually a sane behaviour for updates? mock some scenarios
TODO: correctly implemented the planned behaviour and write tests

## packstate.json

To keep track of changes, mrpack-install creates a `packstate.json` file to store information about the state of a
modpack installation.

#### Example

```json
{
  "slug": "my-cool-pack",
  "project-id": "TDM2I9Mu",
  "version": "1.3.2",
  "version-id": "TpN85eu2",
  "dependencies": {
    "minecraft": "1.19.2",
    "fabric-loader": "0.14.10"
  },
  "hashes": {
    "INFOS.txt": {
      "sha1": "8548257fd702d58e02bd58f6bcc8affbb96a5d38",
      "sha512": "cf83e1357eefb8bdf1542850d66d8007d621e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec3f63b931bd47417a81a538327af927da3e"
    },
    "mods/ultra-blocks-1.2.3+mc1.19.2.jar": {
      "sha1": "df8e2fbabdbeea9c5acea5b1178ad3bb3d75fcf7",
      "sha512": "9884fcf6e1f10c1a6e85fd41515bc2e05390c160dd9c70844d1814a0085dd42c850129b5bf175f6d89ca40b86fbe6407dc3b9a48fb8f393fbd68eafa3b9c5b8c"
    }
  }
}
```
