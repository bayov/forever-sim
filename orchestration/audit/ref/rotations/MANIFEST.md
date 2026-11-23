# Rotation audit: reference files

- `baseline_5518f48_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  5518f4873 (pass 73), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/5518f4873f37/wowsimcli`). Moves since
  7a6d1beb9: Frost +0.53% (835.1 -> 839.6) and Frostfire +0.28% (1006.9 -> 1009.7) from the Ice Lance re-tune (b8b89d6a8);
  nothing else moved.
## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
