# Rotation audit: reference files

- `baseline_70ea588_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  70ea588bb (pass 66). Only the six druid rows were re-simmed (shared CLI `/tmp/audit-cache/cli/70ea588bb116/wowsimcli`):
  0cc2c3c8b (an avoided Ferocious Bite refunds 80%; Sharpened Claws adds spell crit) moved Feral +2.35 (1152.74 ->
  1155.09) and the FF duty +2.36; Balance unchanged. Every other row is carried from the 3a86d4a baseline, since no
  other sim code changed between 3a86d4a6b and 70ea588bb.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
