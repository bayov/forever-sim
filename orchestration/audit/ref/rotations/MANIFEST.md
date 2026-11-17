# Rotation audit: reference files

- `baseline_5f20dfd_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  5f20dfdf4 (pass 68), all 68 rows re-simmed (shared CLI `/tmp/audit-cache/cli/5f20dfdf4cc7/wowsimcli`). Moves since
  1a85df1d7: warriors +3 to +11 (push 114097, Unbridled Wrath on abilities), hunters -20 to -76 (pet white swings
  without the hunter's crit), Subtlety +5.9 (Death Mark always hits).

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
