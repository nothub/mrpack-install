---
id: 10
status: done
blocked-by: [02, 09]
---
# E2E stack loads labrinth's seed data and test PATs

**Delivers:** The e2e stack has labrinth's loaders, game versions and users, plus a user PAT and an admin PAT for tests.

- [x] The staging seed dump is vendored under the e2e package's testdata, from the commit in the pinned image's `org.opencontainers.image.revision` label. It loads after migrations and before labrinth serves
- [x] SQL inserts a user PAT and an admin PAT with the all-scopes mask for that labrinth version. Labrinth treats unknown bits as no scopes
- [x] Tracer test: `GetLatestGameVersion` through `api.NewClient(<labrinth URL>)` returns a seeded version, and `GET /v2/user` with the user PAT answers 200

## Answer
The image revision `8b1d921` holds `apps/labrinth/fixtures/labrinth-seed-data-202508052143.sql`. It is vendored unchanged as `e2e/testdata/` (SHA-1 `032db6a0…`, matches upstream). `seed()` in `e2e/main_test.go` copies it into the postgres container. It then runs one `psql -q -v ON_ERROR_STOP=1 -f <dump> -c <PAT insert>` between migrate and labrinth, and puts psql's output in the error on a non-zero exit. The PATs are `mrp_e2euser` and `mrp_e2eadmin` (constants `userPAT` and `adminPAT`). labrinth stores `access_token` in plain text and routes any token with the `mrp_` prefix to the PAT lookup, so a plain insert works. `e2e/images_test.go` holds every value tied to the revision next to the image ref: `seedDump`, `allScopes = 1<<48 - 1` (bits 0 to 47 of `Scopes` in `pats.rs`), `seedUserID` and `seedAdminID`. One comment there says to re-check them on a bump, which also covers ticket 16's third bullet. `e2e/seed_test.go` asserts that `GetLatestGameVersion` returns `1.21.8`, the newest release in the dump. It also asserts that `GET /v2/user` answers 200 with role `developer` for the user PAT and role `admin` for the admin PAT. The admin check goes beyond the ticket and is kept, because nothing else proves the admin PAT that ticket 11's approval step needs.

Verified: with a 49-bit mask, both PATs get 401, which confirms that unknown bits mean no scopes (`Scopes::from_postgres` falls back to `NONE`). An injected FK failure in the PAT insert surfaced the psql error and left no container, volume or network behind (Ryuk disabled). Readiness went from 14.6 to 19.5 s with seeding included.

Surprises:
- The dump is not purely additive. It `TRUNCATE`s `loaders` (cascading to four join tables) and re-inserts them, then loads game versions as `loader_field_enum_values` (enum 2) after its own `COMMIT`. It must therefore run after migrations, and it cannot run inside an outer `psql -1` transaction.
- The dump also inserts sessions `mra_admin` and `mra_user` for the same two users, expiring 2030-11-03. They are not used: the spec asks for PATs, and the 2100 PAT expiry avoids a time bomb. The PAT comment says so.

Review: both axes approved. Fixed: the seed user IDs moved next to `seedDump`, so a bump checks every value that depends on the dump. The `io.ReadAll` error is now reported instead of dropped, and the exec error reads `exec psql`. Nothing deferred.
