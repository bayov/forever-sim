# Merging PRs into master (B12)

Only when the user asks, one PR at a time.

1. **Order.** Shared sim changes first (the `-consumes` branch), then coordination, then the
   class PRs, starting with those that touch the least shared code.
2. **Merge commits**, not squash or rebase: every class branch carries the shared commits,
   and their IDs must survive for the class diffs to shrink.
3. **Check each class PR's reach** before merging:
   `git diff --name-only origin/master...origin/<branch> | grep -v -E "^sim/<class>/|^ui/<class>|\.results$|^forever-wiki/|^orchestration/"`
   Read every `sim/core`, `sim/common`, `proto/common.proto` and `tools/database` hunk and
   confirm it is gated (`Env.IsForever()` plus a class-only flag or caller). Tell the user
   about anything that changes another class's numbers or the user's own presets and gear.
4. **Stray files.** Look for files at the repo root (built binaries such as `rotopt`, audit
   notes) and large files. Have the owning worker remove them first.
5. **After each merge, fix the next branch.** Its worker merges master in and:
   - regenerates `assets/database/db.json` / `db.bin` with
     `go run ./tools/database/gen_db -outDir=assets -gen=db` (never hand-merge them);
   - regenerates goldens: run the tests (they write `*.results.tmp`), then
     `make update-tests`, and checks that only the intended values moved;
   - runs `go build ./...`, which catches a helper both branches added;
   - keeps both entries when two classes edited the same `tools/database/overrides.go` line.
6. **Test before you merge or say "passes"**: there is no CI on PRs.
   - `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web)` (`sim/web` needs a
     built `binary_dist`).
   - A fresh worktree first needs
     `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`.
