# Rotation audit: reference files

- `baseline_ab3fff7_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  ab3fff7007 (pass 82), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/ab3fff700700/wowsimcli`). Moves since
  8569165 are mostly Phase 0 re-rank gear; the only sim move is +0.8 to +1.6 on the Blackblade paladin builds (rotations-20261126-3).

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
