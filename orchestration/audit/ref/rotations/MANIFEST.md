# Rotation audit: reference files

- `baseline_707ae92_3000it_seed424242.tsv`: DPS of every `tools/dps_report/req/*.json` preset on master
  707ae92cc (pass 59). Against the previous ca5e35c baseline: Enhancement -113 to -123 (Nov 10 shaman tuning
  cb651085c and re-optimized builds 54a15ba24), Affliction +34 to +37 (hotfix 113766 rotation 5607180c7,
  ghost-tick fix f20d5938a), natural-flask presets (warrior, rogue, hunter, Feral) -16 to -28 (70862 flasks
  032692409; Fury -37, SV -27 also take item and hunter mana changes), Arcane -8 (Arcane Concentration split),
  Fire -4; the rest under 0.5.
  - It was made with `tools/audit/rotations/baseline.py`: 3,000 iterations, seed 424242, each req's own
    encounter.
  - The CLI was the audit loop's shared build at 9ba00c2dc (`/tmp/audit-cache/cli/9ba00c2dc473/wowsimcli`).
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
