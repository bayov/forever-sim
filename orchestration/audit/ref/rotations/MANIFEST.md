# Rotation audit: reference files

- `baseline_97da1cc_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  97da1cca5 (pass 63). Against the previous 71562a8 baseline: Retribution -12 to -22 (Unsealed Wrath stores
  no echo from Windfury Totem swings, a1a1a81bb; partly offset by the Ret re-tune ec8ab9b3d), Ret Shock
  +2.4 / +3.5, Combat Swords -3.0 (extra attacks no longer roll Windfury Totem); the rest under 1.
  (The shared CLI was `/tmp/audit-cache/cli/97da1cca5c34/wowsimcli`; the notes below on how it was made
  still hold.)
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at a0b4d3be9 (`/tmp/audit-cache/cli/a0b4d3be9859/wowsimcli`).
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
