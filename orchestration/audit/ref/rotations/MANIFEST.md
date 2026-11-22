# Rotation audit: reference files

- `baseline_7a6d1be_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  7a6d1beb9 (pass 72), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/7a6d1beb9d03/wowsimcli`). Moves since
  8b5137c6d: Frost -11.7% / Frostfire -5.5% (Ice Lance coefficient 0.11), Enhancement -1.0 to -1.1% (Maelstrom off the
  totem, Flame Shock hasted GCD), warriors Fury +0.9 to +1.8% and Arms +0.2 to +0.3% (Deep Wounds, Execute, totem rules),
  Affliction +1.2% (Wrack half-point ticks), hunters -0.1 to -0.4% (pet AP fixed at summon), paladins within ±0.15% (JoW).
## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
