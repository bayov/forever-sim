# Rotation audit: reference files

- `baseline_5a6e9c6_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  5a6e9c6a6 (pass 45). Since c23ef04: the October 30 hotfix talent trees for paladin and druid, and the
  re-searched paladin, druid and mage presets (talents, APLs and reqs together). Ret twist / sub-Holy +59 to
  +83, Ret Shock +15, deep Holy Shock -54 to -62 (Shock vs Undead -22), feral +55, balance +43 to +49,
  Arcane +45, Frost +15. Hunters, warriors and the rest unchanged.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at 5a6e9c6a6 (`/tmp/audit-cache/cli/5a6e9c6a6c2d/wowsimcli`).
  - Next pass: build the CLI on the new master and run `baseline.py` with the same arguments. Diff the
    output against this file. If no req file changed, any row that moved more than about 1 DPS is a sim
    change.
  - When the req files change too, use `regress.py` (old CLI vs new CLI on the new req files) to tell sim
    changes apart from setting changes.

## Not committed (kept locally, re-create as needed)

- Built CLIs. Re-create with
  `AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> tools/audit/rotations/build_cli.sh cli_<sha>`.
  - Keep the previous pass's CLI for `regress.py`. To rebuild an old one, use
    `git worktree add --detach <dir> <old-sha>`, then run `build_cli.sh` with `AUDIT_TREE=<dir>`.
- Generated Go protos (`sim/core/proto/*.pb.go`). `build_cli.sh` regenerates them with protoc.
- No client db2 tables are used by this audit.
