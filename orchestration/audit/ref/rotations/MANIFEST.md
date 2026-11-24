# Rotation audit: reference files

- `baseline_eceb090_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  eceb09019 (pass 75), all 68 rows, 3000 it, seed 424242 (CLI `/tmp/audit-cache/cli/eceb090196be/wowsimcli`). Moves since
  566067a39 come mostly from the Phase 0 regear (step 4) of the req files: warlock -11 to -20%, mage, balance, elemental,
  Fury 2H -12 to -16%, feral, Fury, paladin, shadow, smite -8 to -11%, enhancement, arms -6 to -8%; rogue and hunter unmoved.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
