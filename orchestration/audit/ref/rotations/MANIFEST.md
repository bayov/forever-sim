# Rotation audit: reference files

- `baseline_a0b3069_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  a0b306984 (pass 48). Since 796495d: Frostfire -10.9 (Ice Lance shares Fire Blast's category 19 cooldown, d7541708d),
  the warlocks' Bane of Doom hold (doomAlign, 58a1c3593: Affliction +2.8, duty-curse +3.2, duty-recklessness +1.6,
  Demonology +1.2, Destruction +1.6) and Affliction as an Orc (b33ad307d). Nothing else moved more than 0.5 DPS.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at a0b306984 (`/tmp/audit-cache/cli/a0b306984df7/wowsimcli`).
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
