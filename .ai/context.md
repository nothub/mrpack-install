# mrpack-install

Installs Modrinth modpacks (`.mrpack`) and Minecraft servers into a server dir.

## Language

### Packs

**Override layer**:
One of the `.mrpack` directories whose files are copied verbatim into the server dir: `overrides/` (common), `server-overrides/` and `client-overrides/`. mrpack-install applies `overrides/` first and `server-overrides/` over it, and skips `client-overrides/`.
_Avoid_: override folder, override set

### E2E tests

**labrinth**:
Modrinth's backend server, run from its container image as the API under test. Lowercase, as upstream writes it.
_Avoid_: Modrinth API (for the local instance), backend, mock server

**Stack**:
The set of containers the e2e `TestMain` boots once per run: labrinth, its backing services, and the CDN.
_Avoid_: environment, compose, cluster

**Seed dump**:
labrinth's own SQL fixture file, vendored from the image's revision, that loads base data (loaders, game versions, users) before labrinth starts. Projects and versions are not part of it.
_Avoid_: seed data (ambiguous with API-created fixtures), fixtures, DB snapshot
