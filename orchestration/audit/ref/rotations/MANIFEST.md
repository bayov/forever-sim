# Rotation audit: reference files

- `baseline_f5c99ad_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  f5c99ad01 (pass 49, round 70753: client 1.60.1.70753 and the 2026-11-02 dev notes, every class reworked).
  Against a0b3069 every class but priest and mage moved by tens to hundreds of DPS (warrior +76..+127, rogue
  +122..+209, shaman +201..+275, hunter +16..+224); the six rogue "-humanoid" presets are gone and
  `warlock__affliction_wrack` is new. Attribute those moves to the round files (`orchestration/round_70753/`).
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at f5c99ad01 (`/tmp/audit-cache/cli/f5c99ad018aa/wowsimcli`).
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
