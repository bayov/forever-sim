# Rotation audit: reference files

- `baseline_566067a_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  566067a39 (pass 74), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/566067a39465/wowsimcli`). Moves since
  5518f4873: feral cat -15.0 and -14.9 (Rip/Rake AP shares, be773f5eb); Fury -3.6, duty-battle-shout -5.5, duty-sunder -1.5,
  Fury 2H -0.6 / -1.3 (warrior proc rework, cd71e78c2); BM +1.2 to +1.4 and MM +0.7 to +0.8 (pet summoned with Hawk, b82ad9a1e).
## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
